package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"

	"google.golang.org/protobuf/proto"
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

// RoleStatus 展示当前角色在服务器中的状态
type RoleStatus struct {
	Account      string
	RoleID       uint64
	RoleName     string
	Level        uint32
	SceneID      uint32
	EntityHandle uint64
	PosX         uint32
	PosY         uint32
	HP           int64
	MP           int64
	StateFlags   uint64
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

	dataMu       sync.RWMutex
	roleID       uint64
	roleName     string
	roleLevel    uint32
	sceneID      uint32
	entityHandle uint64
	posX         uint32
	posY         uint32
	stateFlags   uint64
	hp           int64
	mp           int64

	observedMu sync.RWMutex
	observed   map[uint64]*EntityView

	syncMu   sync.RWMutex
	timeSync struct {
		serverMs int64
		localMs  int64
	}

	flow struct {
		registerCh    chan *protocol.S2CRegisterResultReq
		loginCh       chan *protocol.S2CLoginResultReq
		roleListCh    chan *protocol.S2CRoleListReq
		createRoleCh  chan *protocol.S2CCreateRoleResultReq
		enterSceneCh  chan *protocol.S2CEnterSceneReq
		aoiCh         chan *EntityView
		skillDamageCh chan *protocol.S2CSkillDamageResultReq
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
	client.flow.skillDamageCh = make(chan *protocol.S2CSkillDamageResultReq, 4)
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

func (c *GameClient) RegisterAccount(username, password string) error {
	c.username = username
	c.password = password
	req := &protocol.C2SRegisterReq{
		Username: username,
		Password: password,
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

func (c *GameClient) LoginAccount(username, password string) error {
	c.username = username
	c.password = password
	req := &protocol.C2SLoginReq{
		Username: username,
		Password: password,
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

func (c *GameClient) ListRoles() ([]*protocol.PlayerSimpleData, error) {
	if err := c.QueryRoles(); err != nil {
		return nil, err
	}
	resp, err := waitForResponse(c.flow.roleListCh, defaultClientTimeout)
	if err != nil {
		return nil, err
	}
	return resp.RoleList, nil
}

func (c *GameClient) QueryRoles() error {
	return c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SQueryRoles), &protocol.C2SQueryRolesReq{})
}

func (c *GameClient) CreateRole(roleName string, job, sex uint32) error {
	req := &protocol.C2SCreateRoleReq{
		RoleData: &protocol.PlayerSimpleData{
			RoleName: roleName,
			Job:      job,
			Sex:      sex,
		},
	}
	if err := c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SCreateRole), req); err != nil {
		return err
	}
	_, err := waitForResponse(c.flow.createRoleCh, defaultClientTimeout)
	if err == nil {
		c.dataMu.Lock()
		c.roleName = roleName
		c.dataMu.Unlock()
	}
	return err
}

func (c *GameClient) EnterGame(roleID uint64) error {
	req := &protocol.C2SEnterGameReq{RoleId: roleID}
	if err := c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SEnterGame), req); err != nil {
		return err
	}
	if _, err := waitForResponse(c.flow.enterSceneCh, defaultClientTimeout); err != nil {
		return err
	}
	c.dataMu.Lock()
	c.roleID = roleID
	c.dataMu.Unlock()
	return nil
}

func (c *GameClient) NudgeMove(dx, dy int32) error {
	c.dataMu.RLock()
	startX := c.posX
	startY := c.posY
	c.dataMu.RUnlock()

	targetX := clampToUint32(int64(startX) + int64(dx))
	targetY := clampToUint32(int64(startY) + int64(dy))
	seq := uint32(servertime.Now().UnixNano() & 0xffffffff)

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
	c.dataMu.RLock()
	posX := c.posX
	posY := c.posY
	c.dataMu.RUnlock()

	req := &protocol.C2SUseSkillReq{
		SkillId:   0,
		TargetHdl: targetHdl,
		PosX:      posX,
		PosY:      posY,
	}
	return c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SUseSkill), req)
}

func (c *GameClient) WaitForEntityInView(timeout time.Duration) (*EntityView, error) {
	return waitForResponse(c.flow.aoiCh, timeout)
}

func (c *GameClient) WaitForSkillResult(targetHdl uint64, timeout time.Duration) (*protocol.SkillHitResultSt, error) {
	deadline := servertime.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, fmt.Errorf("[%s] wait skill result timeout", c.id)
		}

		timer := time.NewTimer(remaining)
		select {
		case resp := <-c.flow.skillDamageCh:
			timer.Stop()
			for _, hit := range resp.Hits {
				if hit.TargetHdl == targetHdl {
					return hit, nil
				}
			}
		case <-timer.C:
			return nil, fmt.Errorf("[%s] wait skill result timeout", c.id)
		}
	}
}

func (c *GameClient) ObservedEntities() []*EntityView {
	c.observedMu.RLock()
	defer c.observedMu.RUnlock()
	results := make([]*EntityView, 0, len(c.observed))
	for _, view := range c.observed {
		copyView := *view
		results = append(results, &copyView)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Handle < results[j].Handle
	})
	return results
}

func (c *GameClient) RoleStatus() RoleStatus {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	return RoleStatus{
		Account:      c.username,
		RoleID:       c.roleID,
		RoleName:     c.roleName,
		Level:        c.roleLevel,
		SceneID:      c.sceneID,
		EntityHandle: c.entityHandle,
		PosX:         c.posX,
		PosY:         c.posY,
		HP:           c.hp,
		MP:           c.mp,
		StateFlags:   c.stateFlags,
	}
}

func (c *GameClient) HasEnteredScene() bool {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	return c.entityHandle != 0
}

func (c *GameClient) AccountName() string {
	return c.username
}

func (c *GameClient) LastServerTime() (int64, bool) {
	c.syncMu.RLock()
	defer c.syncMu.RUnlock()
	if c.timeSync.serverMs == 0 {
		return 0, false
	}
	delta := servertime.Now().UnixMilli() - c.timeSync.localMs
	return c.timeSync.serverMs + delta, true
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
	c.sceneID = entity.SceneId
	c.roleLevel = entity.Level
	if entity.ShowName != "" {
		c.roleName = entity.ShowName
	}
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
	if resp.ErrCode != 0 {
		log.Warnf("[%s] skill cast failed, skillId=%d err=%d", c.id, resp.SkillId, resp.ErrCode)
	}
}

func (c *GameClient) OnSkillDamageResult(resp *protocol.S2CSkillDamageResultReq) {
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
	case c.flow.skillDamageCh <- resp:
	default:
	}
}

func (c *GameClient) OnTimeSync(resp *protocol.S2CTimeSyncReq) {
	localMs := servertime.Now().UnixMilli()
	c.syncMu.Lock()
	c.timeSync.serverMs = resp.ServerTimeMs
	c.timeSync.localMs = localMs
	c.syncMu.Unlock()

	diff := resp.ServerTimeMs - localMs
	if diff < 0 {
		diff = -diff
	}
	if diff > 200 {
		log.Warnf("[%s] server time drift detected: %dms", c.id, diff)
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
	deadline := servertime.Now().Add(timeout)
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()
	select {
	case resp := <-ch:
		return resp, nil
	case <-timer.C:
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
