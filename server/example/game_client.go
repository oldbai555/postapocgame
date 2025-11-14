package main

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"sync"
	"time"
)

const defaultClientTimeout = 10 * time.Second

// EntityView 记录视野内的实体快照
type EntityView struct {
	Handle     uint64
	PosX       uint32
	PosY       uint32
	Hp         int64
	HasHp      bool
	Mp         int64
	HasMp      bool
	StateFlags uint64
}

// GameClient 游戏客户端（使用Actor）
type GameClient struct {
	id          string
	gatewayAddr string
	tcpClient   network.ITCPClient
	codec       *network.Codec
	actorMgr    actor.IActorManager
	actorCtx    actor.IActorContext

	username string
	password string
	roleName string

	dataMu       sync.RWMutex
	roleID       uint64
	entityHandle uint64
	posX         uint32
	posY         uint32
	stateFlags   uint64
	hp           int64
	mp           int64

	observedMu sync.RWMutex
	observed   map[uint64]*EntityView

	flow struct {
		registerCh   chan *protocol.S2CRegisterResultReq
		loginCh      chan *protocol.S2CLoginResultReq
		roleListCh   chan *protocol.S2CRoleListReq
		createRoleCh chan *protocol.S2CCreateRoleResultReq
		enterSceneCh chan *protocol.S2CEnterSceneReq
		aoiCh        chan *EntityView
		skillCh      chan *protocol.S2CSkillCastResultReq
	}
}

func NewGameClient(playerID string, gatewayAddr string, actorMgr actor.IActorManager) *GameClient {
	client := &GameClient{
		id:          playerID,
		gatewayAddr: gatewayAddr,
		codec:       network.DefaultCodec(),
		actorMgr:    actorMgr,
		observed:    make(map[uint64]*EntityView),
	}
	client.flow.registerCh = make(chan *protocol.S2CRegisterResultReq, 1)
	client.flow.loginCh = make(chan *protocol.S2CLoginResultReq, 1)
	client.flow.roleListCh = make(chan *protocol.S2CRoleListReq, 1)
	client.flow.createRoleCh = make(chan *protocol.S2CCreateRoleResultReq, 1)
	client.flow.enterSceneCh = make(chan *protocol.S2CEnterSceneReq, 1)
	client.flow.aoiCh = make(chan *EntityView, 4)
	client.flow.skillCh = make(chan *protocol.S2CSkillCastResultReq, 4)
	return client
}

// Start 连接服务器
func (c *GameClient) Start(ctx context.Context) error {
	handler := &NetworkMessageHandler{client: c}
	c.tcpClient = network.NewTCPClient(
		network.WithTCPClientOptionNetworkMessageHandler(handler),
		network.WithTCPClientOptionOnDisConn(func(conn network.IConnection) {
			log.Infof("[%s] disconnected from gateway", c.id)
		}),
		network.WithTCPClientOptionOnConn(func(conn network.IConnection) {
			log.Infof("[%s] connected to gateway", c.id)
		}),
	)

	if err := c.tcpClient.Connect(ctx, c.gatewayAddr); err != nil {
		return fmt.Errorf("connect gateway failed: %w", err)
	}

	actorCtx, err := c.actorMgr.GetOrCreateActor(c.id)
	if err != nil {
		return customerr.Wrap(err)
	}
	c.actorCtx = actorCtx
	c.actorCtx.SetData("gameClient", c)
	return nil
}

func (c *GameClient) Close() {
	if c.tcpClient != nil {
		_ = c.tcpClient.Close()
	}
	if c.actorMgr != nil && c.id != "" {
		_ = c.actorMgr.RemoveActor(c.id)
	}
}

func (c *GameClient) sendProtoMessage(msgID uint16, payload proto.Message) error {
	data, err := proto.Marshal(payload)
	if err != nil {
		return customerr.Wrap(err)
	}
	bytes, err := c.codec.EncodeClientMessage(&network.ClientMessage{
		MsgId: msgID,
		Data:  data,
	})
	if err != nil {
		return customerr.Wrap(err)
	}
	conn := c.tcpClient.GetConnection()
	if conn == nil {
		return errors.New("connection closed")
	}
	return conn.SendMessage(&network.Message{
		Type:    network.MsgTypeClient,
		Payload: bytes,
	})
}

// RunLoginFlow 按照测试流程执行注册、登录、建角、进副本
func (c *GameClient) RunLoginFlow(identity TestIdentity) error {
	c.username = identity.Account
	c.password = identity.Password
	c.roleName = identity.RoleName

	if err := c.RegisterAccount(); err != nil {
		log.Warnf("[%s] register failed (maybe already exists): %v", c.id, err)
	}
	if err := c.LoginAccount(); err != nil {
		return err
	}
	roleID, err := c.EnsureRole()
	if err != nil {
		return err
	}
	c.roleID = roleID
	return c.EnterGame(roleID)
}

func (c *GameClient) RegisterAccount() error {
	req := &protocol.C2SRegisterReq{
		Username: c.username,
		Password: c.password,
	}
	if err := c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SRegister), req); err != nil {
		return err
	}
	resp, err := waitForResponse(c.flow.registerCh, defaultClientTimeout)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("register failed: %s", resp.Message)
	}
	return nil
}

func (c *GameClient) LoginAccount() error {
	req := &protocol.C2SLoginReq{
		Username: c.username,
		Password: c.password,
	}
	if err := c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SLogin), req); err != nil {
		return err
	}
	resp, err := waitForResponse(c.flow.loginCh, defaultClientTimeout)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("login failed: %s", resp.Message)
	}
	return nil
}

func (c *GameClient) EnsureRole() (uint64, error) {
	if err := c.QueryRoles(); err != nil {
		return 0, err
	}
	roleList, err := waitForResponse(c.flow.roleListCh, defaultClientTimeout)
	if err != nil {
		return 0, err
	}
	if len(roleList.RoleList) == 0 {
		if err := c.CreateRole(); err != nil {
			return 0, err
		}
		if err := c.QueryRoles(); err != nil {
			return 0, err
		}
		roleList, err = waitForResponse(c.flow.roleListCh, defaultClientTimeout)
		if err != nil {
			return 0, err
		}
	}
	c.roleID = roleList.RoleList[0].RoleId
	return c.roleID, nil
}

func (c *GameClient) QueryRoles() error {
	return c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SQueryRoles), &protocol.C2SQueryRolesReq{})
}

func (c *GameClient) CreateRole() error {
	req := &protocol.C2SCreateRoleReq{
		RoleData: &protocol.PlayerSimpleData{
			RoleName: c.roleName,
			Job:      1,
			Sex:      1,
		},
	}
	if err := c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SCreateRole), req); err != nil {
		return err
	}
	_, err := waitForResponse(c.flow.createRoleCh, defaultClientTimeout)
	return err
}

func (c *GameClient) EnterGame(roleID uint64) error {
	req := &protocol.C2SEnterGameReq{RoleId: roleID}
	if err := c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SEnterGame), req); err != nil {
		return err
	}
	_, err := waitForResponse(c.flow.enterSceneCh, defaultClientTimeout)
	return err
}

func (c *GameClient) NudgeMove(dx, dy int32) error {
	c.dataMu.RLock()
	startX := c.posX
	startY := c.posY
	c.dataMu.RUnlock()

	targetX := clampToUint32(int64(startX) + int64(dx))
	targetY := clampToUint32(int64(startY) + int64(dy))
	seq := uint32(time.Now().UnixNano() & 0xffffffff)

	startReq := &protocol.C2SStartMoveReq{
		FromX: startX,
		FromY: startY,
		ToX:   targetX,
		ToY:   targetY,
		Speed: 480,
		Seq:   seq,
	}
	if err := c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SStartMove), startReq); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	endReq := &protocol.C2SEndMoveReq{
		PosX: targetX,
		PosY: targetY,
		Seq:  seq + 1,
	}
	return c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SEndMove), endReq)
}

func (c *GameClient) CastNormalAttack(targetHdl uint64) error {
	req := &protocol.C2SUseSkillReq{
		SkillId:   0,
		TargetHdl: targetHdl,
		PosX:      c.posX,
		PosY:      c.posY,
	}
	return c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SUseSkill), req)
}

func (c *GameClient) WaitForEntityInView(timeout time.Duration) (*EntityView, error) {
	return waitForResponse(c.flow.aoiCh, timeout)
}

func (c *GameClient) WaitForSkillResult(targetHdl uint64, timeout time.Duration) (*protocol.SkillHitResultSt, error) {
	deadline := time.After(timeout)
	for {
		select {
		case resp := <-c.flow.skillCh:
			for _, hit := range resp.Hits {
				if hit.TargetHdl == targetHdl {
					return hit, nil
				}
			}
		case <-deadline:
			return nil, fmt.Errorf("[%s] wait skill result timeout", c.id)
		}
	}
}

func (c *GameClient) EntityHandle() uint64 {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	return c.entityHandle
}

func (c *GameClient) GetPlayerID() string {
	return c.id
}

// ---- handlers ----

func (c *GameClient) OnRegisterResult(resp *protocol.S2CRegisterResultReq) {
	select {
	case c.flow.registerCh <- resp:
	default:
	}
}

func (c *GameClient) OnLoginResult(resp *protocol.S2CLoginResultReq) {
	select {
	case c.flow.loginCh <- resp:
	default:
	}
}

func (c *GameClient) OnRoleList(resp *protocol.S2CRoleListReq) {
	select {
	case c.flow.roleListCh <- resp:
	default:
	}
}

func (c *GameClient) OnCreateRoleResult(resp *protocol.S2CCreateRoleResultReq) {
	select {
	case c.flow.createRoleCh <- resp:
	default:
	}
}

func (c *GameClient) OnEnterScene(resp *protocol.S2CEnterSceneReq) {
	entity := resp.EntityData
	c.dataMu.Lock()
	c.entityHandle = entity.Hdl
	c.posX = entity.PosX
	c.posY = entity.PosY
	c.hp = attrValueOrZero(entity.Attrs, attrdef.AttrHP)
	c.mp = attrValueOrZero(entity.Attrs, attrdef.AttrMP)
	c.stateFlags = entity.StateFlags
	c.dataMu.Unlock()

	select {
	case c.flow.enterSceneCh <- resp:
	default:
	}
}

func (c *GameClient) OnEntityMove(resp *protocol.S2CEntityMoveReq) {
	if resp.EntityHdl == c.EntityHandle() {
		c.dataMu.Lock()
		c.posX = resp.PosX
		c.posY = resp.PosY
		c.dataMu.Unlock()
		return
	}
	view := &EntityView{
		Handle: resp.EntityHdl,
		PosX:   resp.PosX,
		PosY:   resp.PosY,
	}
	c.updateObserved(view)
	select {
	case c.flow.aoiCh <- view:
	default:
	}
}

func (c *GameClient) OnEntityStop(resp *protocol.S2CEntityStopMoveReq) {
	c.OnEntityMove(&protocol.S2CEntityMoveReq{
		EntityHdl: resp.EntityHdl,
		PosX:      resp.PosX,
		PosY:      resp.PosY,
	})
}

func (c *GameClient) OnSkillCastResult(resp *protocol.S2CSkillCastResultReq) {
	for _, hit := range resp.Hits {
		view := &EntityView{
			Handle:     hit.TargetHdl,
			StateFlags: hit.StateFlags,
		}
		if hp, ok := attrValue(hit.Attrs, attrdef.AttrHP); ok {
			view.Hp = hp
			view.HasHp = true
		}
		if mp, ok := attrValue(hit.Attrs, attrdef.AttrMP); ok {
			view.Mp = mp
			view.HasMp = true
		}
		c.updateObserved(view)
	}
	select {
	case c.flow.skillCh <- resp:
	default:
	}
}

func (c *GameClient) updateObserved(view *EntityView) {
	c.observedMu.Lock()
	defer c.observedMu.Unlock()
	if existing, ok := c.observed[view.Handle]; ok {
		if view.HasHp {
			existing.Hp = view.Hp
			existing.HasHp = true
		}
		if view.HasMp {
			existing.Mp = view.Mp
			existing.HasMp = true
		}
		if view.StateFlags != 0 {
			existing.StateFlags = view.StateFlags
		}
		if view.PosX != 0 || view.PosY != 0 {
			existing.PosX = view.PosX
			existing.PosY = view.PosY
		}
	} else {
		c.observed[view.Handle] = &EntityView{
			Handle:     view.Handle,
			PosX:       view.PosX,
			PosY:       view.PosY,
			StateFlags: view.StateFlags,
			Hp:         view.Hp,
			Mp:         view.Mp,
			HasHp:      view.HasHp,
			HasMp:      view.HasMp,
		}
	}
}

// waitForResponse 等待带超时的响应
func waitForResponse[T any](ch <-chan T, timeout time.Duration) (T, error) {
	var zero T
	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(timeout):
		return zero, fmt.Errorf("wait response timeout (%s)", timeout)
	}
}

func clampToUint32(v int64) uint32 {
	if v < 0 {
		return 0
	}
	if v > int64(^uint32(0)) {
		return ^uint32(0)
	}
	return uint32(v)
}

func attrValue(attrs map[uint32]int64, attrType attrdef.AttrType) (int64, bool) {
	if attrs == nil {
		return 0, false
	}
	val, ok := attrs[uint32(attrType)]
	return val, ok
}

func attrValueOrZero(attrs map[uint32]int64, attrType attrdef.AttrType) int64 {
	val, _ := attrValue(attrs, attrType)
	return val
}
