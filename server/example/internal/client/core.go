package client

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
)

const (
	defaultClientTimeout = 10 * time.Second
	defaultMoveSpeed     = 480
	defaultMoveStep      = 64
	moveUpdateInterval   = 80 * time.Millisecond
)

// EntityView 记录视野内的实体快照
// 注意：PosX、PosY 是格子坐标（不是像素坐标）
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
// 注意：PosX、PosY 是格子坐标（不是像素坐标）
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

// Core 表示调试客户端的核心能力（连接、状态、协议流）
type Core struct {
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
	sceneMap     *jsonconf.GameMap
	moveRunner   *MoveRunner

	bagMu    sync.RWMutex
	bagItems []*protocol.ItemSt

	moneyMu   sync.RWMutex
	moneyData map[uint32]int64

	observedMu sync.RWMutex
	observed   map[uint64]*EntityView

	syncMu   sync.RWMutex
	timeSync struct {
		serverMs int64
		localMs  int64
	}

	flow flowRegistry
}

func NewCore(playerID string, gatewayAddr string, actorMgr actor.IActorManager) *Core {
	core := &Core{
		id:          playerID,
		gatewayAddr: gatewayAddr,
		codec:       network.DefaultCodec(),
		actorMgr:    actorMgr,
		observed:    make(map[uint64]*EntityView),
		flow:        newFlowRegistry(),
		moneyData:   make(map[uint32]int64),
	}
	core.moveRunner = NewMoveRunner(core)
	return core
}

// Start 连接 Gateway 并创建 Actor
func (c *Core) Start(ctx context.Context) error {
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

func (c *Core) Close() {
	if c.tcpClient != nil {
		_ = c.tcpClient.Close()
	}
	if c.actorMgr != nil && c.id != "" {
		_ = c.actorMgr.RemoveActor(c.id)
	}
}

// --- 协议发送 ---

func (c *Core) sendProtoMessage(msgID uint16, payload proto.Message) error {
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

func (c *Core) SendClientProto(msgID uint16, payload proto.Message) error {
	return c.sendProtoMessage(msgID, payload)
}

// --- 账号/角色 ---

func (c *Core) RegisterAccount(username, password string) error {
	c.username = username
	c.password = password
	req := &protocol.C2SRegisterReq{
		Username: username,
		Password: password,
	}
	if err := c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SRegister), req); err != nil {
		return err
	}
	resp, err := c.flow.register.Wait(defaultClientTimeout)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("register failed: %s", resp.Message)
	}
	return nil
}

func (c *Core) LoginAccount(username, password string) error {
	c.username = username
	c.password = password
	req := &protocol.C2SLoginReq{
		Username: username,
		Password: password,
	}
	if err := c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SLogin), req); err != nil {
		return err
	}
	resp, err := c.flow.login.Wait(defaultClientTimeout)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("login failed: %s", resp.Message)
	}
	return nil
}

func (c *Core) ListRoles() ([]*protocol.PlayerSimpleData, error) {
	if err := c.QueryRoles(); err != nil {
		return nil, err
	}
	resp, err := c.flow.roleList.Wait(defaultClientTimeout)
	if err != nil {
		return nil, err
	}
	return resp.RoleList, nil
}

func (c *Core) QueryRoles() error {
	return c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SQueryRoles), &protocol.C2SQueryRolesReq{})
}

func (c *Core) CreateRole(roleName string, job, sex uint32) error {
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
	if _, err := c.flow.createRole.Wait(defaultClientTimeout); err == nil {
		c.dataMu.Lock()
		c.roleName = roleName
		c.dataMu.Unlock()
		return nil
	} else {
		return err
	}
}

func (c *Core) EnterGame(roleID uint64) error {
	req := &protocol.C2SEnterGameReq{RoleId: roleID}
	if err := c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SEnterGame), req); err != nil {
		return err
	}
	if _, err := c.flow.enterScene.Wait(defaultClientTimeout); err != nil {
		return err
	}
	c.dataMu.Lock()
	c.roleID = roleID
	c.dataMu.Unlock()
	return nil
}

// --- 移动与战斗 ---

func (c *Core) NudgeMove(dx, dy int32) error {
	step := int64(defaultMoveStep)
	if step <= 0 {
		step = 64
	}
	if err := c.moveAlongAxis(int64(dx), step, true); err != nil {
		return err
	}
	if err := c.moveAlongAxis(int64(dy), step, false); err != nil {
		return err
	}
	return nil
}

func (c *Core) CastNormalAttack(targetHdl uint64) error {
	c.dataMu.RLock()
	tileX := c.posX
	tileY := c.posY
	c.dataMu.RUnlock()

	pixelX, pixelY := argsdef.TileCoordToPixel(tileX, tileY)
	req := &protocol.C2SUseSkillReq{
		SkillId:   0,
		TargetHdl: targetHdl,
		PosX:      pixelX,
		PosY:      pixelY,
	}
	return c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SUseSkill), req)
}

func (c *Core) WaitForSkillResult(targetHdl uint64, timeout time.Duration) (*protocol.SkillHitResultSt, error) {
	deadline := servertime.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, fmt.Errorf("[%s] wait skill result timeout", c.id)
		}

		timer := time.NewTimer(remaining)
		select {
		case resp := <-c.flow.skillDamage.Chan():
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

// --- 状态查询 ---

func (c *Core) ObservedEntities() []*EntityView {
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

func (c *Core) RoleStatus() RoleStatus {
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

func (c *Core) HasEnteredScene() bool {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	return c.entityHandle != 0
}

func (c *Core) AccountName() string {
	return c.username
}

func (c *Core) LastServerTime() (int64, bool) {
	c.syncMu.RLock()
	defer c.syncMu.RUnlock()
	if c.timeSync.serverMs == 0 {
		return 0, false
	}
	delta := servertime.Now().UnixMilli() - c.timeSync.localMs
	return c.timeSync.serverMs + delta, true
}

func (c *Core) EntityHandle() uint64 {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	return c.entityHandle
}

func (c *Core) SceneID() uint32 {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	return c.sceneID
}

// --- Flow 回调入口 ---

func (c *Core) OnRegisterResult(resp *protocol.S2CRegisterResultReq) {
	c.flow.register.Deliver(resp)
}

func (c *Core) OnLoginResult(resp *protocol.S2CLoginResultReq) {
	c.flow.login.Deliver(resp)
}

func (c *Core) OnRoleList(resp *protocol.S2CRoleListReq) {
	c.flow.roleList.Deliver(resp)
}

func (c *Core) OnCreateRoleResult(resp *protocol.S2CCreateRoleResultReq) {
	c.flow.createRole.Deliver(resp)
}

func (c *Core) OnEnterScene(resp *protocol.S2CEnterSceneReq) {
	entity := resp.EntityData
	sceneMap := c.lookupSceneMap(entity.SceneId)
	if sceneMap == nil {
		log.Warnf("[%s] scene %d has no map data, movement will not be clamped", c.id, entity.SceneId)
	}
	c.dataMu.Lock()
	c.entityHandle = entity.Hdl
	c.posX = entity.PosX
	c.posY = entity.PosY
	c.hp = attrValueOrZero(entity.Attrs, attrdef.AttrHP)
	c.mp = attrValueOrZero(entity.Attrs, attrdef.AttrMP)
	c.stateFlags = entity.StateFlags
	c.sceneID = entity.SceneId
	c.roleLevel = entity.Level
	c.sceneMap = sceneMap
	if entity.ShowName != "" {
		c.roleName = entity.ShowName
	}
	c.dataMu.Unlock()

	c.flow.enterScene.Deliver(resp)
}

func (c *Core) OnEntityMove(resp *protocol.S2CEntityMoveReq) {
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
	case c.flow.aoi.Chan() <- view:
	default:
	}
}

func (c *Core) OnEntityStop(resp *protocol.S2CEntityStopMoveReq) {
	c.OnEntityMove(&protocol.S2CEntityMoveReq{
		EntityHdl: resp.EntityHdl,
		PosX:      resp.PosX,
		PosY:      resp.PosY,
	})
}

func (c *Core) OnStartMove(resp *protocol.S2CStartMoveReq) {
	// currently no-op; kept for protocol completeness
}

func (c *Core) OnEndMove(resp *protocol.S2CEndMoveReq) {
	// currently no-op; kept for protocol completeness
}

func (c *Core) OnSkillCastResult(resp *protocol.S2CSkillCastResultReq) {
	if resp.ErrCode != 0 {
		log.Warnf("[%s] skill cast failed, skillId=%d err=%d", c.id, resp.SkillId, resp.ErrCode)
	}
}

func (c *Core) OnSkillDamageResult(resp *protocol.S2CSkillDamageResultReq) {
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
	case c.flow.skillDamage.Chan() <- resp:
	default:
	}
}

func (c *Core) OnTimeSync(resp *protocol.S2CTimeSyncReq) {
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

func (c *Core) OnBagData(resp *protocol.S2CBagDataReq) {
	if resp == nil || resp.BagData == nil {
		return
	}
	c.bagMu.Lock()
	c.bagItems = c.bagItems[:0]
	for _, item := range resp.BagData.Items {
		if item == nil {
			continue
		}
		copyItem := proto.Clone(item).(*protocol.ItemSt)
		c.bagItems = append(c.bagItems, copyItem)
	}
	c.bagMu.Unlock()
	c.flow.bagData.Deliver(resp)
}

func (c *Core) OnBagUpdate() {
	// 服务端提示背包更新，主动重新拉取
	if err := c.RequestBagData(); err != nil {
		log.Warnf("[%s] request bag data failed: %v", c.id, err)
	}
}

func (c *Core) OnMoneyData(resp *protocol.S2CMoneyDataReq) {
	if resp == nil || resp.MoneyData == nil {
		return
	}
	c.moneyMu.Lock()
	for k, v := range resp.MoneyData.MoneyMap {
		c.moneyData[k] = v
	}
	c.moneyMu.Unlock()
	c.flow.moneyData.Deliver(resp)
}

func (c *Core) OnGMCommandResult(resp *protocol.S2CGMCommandResultReq) {
	c.flow.gmResult.Deliver(resp)
}

func (c *Core) OnUseItemResult(resp *protocol.S2CUseItemResultReq) {
	c.flow.useItem.Deliver(resp)
}

func (c *Core) OnPickupItemResult(resp *protocol.S2CPickupItemResultReq) {
	c.flow.pickup.Deliver(resp)
}

func (c *Core) OnEnterDungeonResult(resp *protocol.S2CEnterDungeonResultReq) {
	c.flow.dungeonEnter.Deliver(resp)
}

// --- 内部辅助 ---

func (c *Core) updateObserved(view *EntityView) {
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

func (c *Core) lookupSceneMap(sceneID uint32) *jsonconf.GameMap {
	sceneCfg, ok := jsonconf.GetConfigManager().GetSceneConfig(sceneID)
	if !ok || sceneCfg == nil {
		return nil
	}
	return sceneCfg.GameMap
}

func (c *Core) CurrentSceneMap() *jsonconf.GameMap {
	c.dataMu.RLock()
	sceneMap := c.sceneMap
	c.dataMu.RUnlock()
	return sceneMap
}

func (c *Core) prepareMoveTarget(targetX, targetY uint32) (uint32, uint32, error) {
	sceneMap := c.CurrentSceneMap()
	if sceneMap == nil {
		return targetX, targetY, nil
	}
	clampedX, clampedY := sceneMap.ClampCoord(int64(targetX), int64(targetY))
	targetX = uint32(clampedX)
	targetY = uint32(clampedY)
	if !sceneMap.IsWalkable(int32(targetX), int32(targetY)) {
		return targetX, targetY, fmt.Errorf("[%s] target (%d,%d) not walkable", c.id, targetX, targetY)
	}
	return targetX, targetY, nil
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

func clampDelta(remaining, limit int64) int64 {
	if remaining > limit {
		return limit
	}
	if remaining < -limit {
		return -limit
	}
	return remaining
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

func sign64(v int64) int64 {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	default:
		return 0
	}
}

func (c *Core) sendMoveChunk(fromX, fromY, toX, toY uint32) error {
	if fromX == toX && fromY == toY {
		return nil
	}
	speed := uint32(defaultMoveSpeed)

	fromPixelX, fromPixelY := argsdef.TileCoordToPixel(fromX, fromY)
	toPixelX, toPixelY := argsdef.TileCoordToPixel(toX, toY)
	startReq := &protocol.C2SStartMoveReq{
		FromX: fromPixelX,
		FromY: fromPixelY,
		ToX:   toPixelX,
		ToY:   toPixelY,
		Speed: speed,
	}
	if err := c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SStartMove), startReq); err != nil {
		return err
	}

	var (
		dx        = int64(toX) - int64(fromX)
		dy        = int64(toY) - int64(fromY)
		stepCount = abs64(dx) + abs64(dy)
		stepX     = sign64(dx)
		stepY     = sign64(dy)
		curX      = int64(fromX)
		curY      = int64(fromY)
	)

	for i := int64(0); i < stepCount; i++ {
		curX += stepX
		curY += stepY
		time.Sleep(moveUpdateInterval)
		updatePixelX, updatePixelY := argsdef.TileCoordToPixel(uint32(curX), uint32(curY))
		updateReq := &protocol.C2SUpdateMoveReq{
			PosX:  updatePixelX,
			PosY:  updatePixelY,
			Speed: speed,
		}
		if err := c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SUpdateMove), updateReq); err != nil {
			return err
		}
	}

	endPixelX, endPixelY := argsdef.TileCoordToPixel(toX, toY)
	endReq := &protocol.C2SEndMoveReq{
		PosX: endPixelX,
		PosY: endPixelY,
	}
	return c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SEndMove), endReq)
}

func (c *Core) moveAlongAxis(delta int64, step int64, isX bool) error {
	remaining := delta
	for remaining != 0 {
		chunk := clampDelta(remaining, step)

		c.dataMu.RLock()
		startX := c.posX
		startY := c.posY
		c.dataMu.RUnlock()

		targetX := startX
		targetY := startY
		if isX {
			targetX = clampToUint32(int64(startX) + chunk)
		} else {
			targetY = clampToUint32(int64(startY) + chunk)
		}

		var err error
		targetX, targetY, err = c.prepareMoveTarget(targetX, targetY)
		if err != nil {
			return err
		}

		if err := c.sendMoveChunk(startX, startY, targetX, targetY); err != nil {
			return err
		}

		c.dataMu.Lock()
		c.posX = targetX
		c.posY = targetY
		c.dataMu.Unlock()

		remaining -= chunk
	}
	return nil
}

func (c *Core) GetPlayerID() string {
	return c.id
}

func (c *Core) GatewayAddr() string {
	return c.gatewayAddr
}

func (c *Core) MoveRunner() *MoveRunner {
	return c.moveRunner
}

func (c *Core) updateLocalPosition(tileX, tileY uint32) {
	c.dataMu.Lock()
	c.posX = tileX
	c.posY = tileY
	c.dataMu.Unlock()
}

func (c *Core) currentPosition() (uint32, uint32) {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	return c.posX, c.posY
}

// --- Bag & Inventory ---

func (c *Core) RequestBagData() error {
	return c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SOpenBag), &protocol.C2SOpenBagReq{})
}

func (c *Core) WaitBagData(timeout time.Duration) (*protocol.S2CBagDataReq, error) {
	return c.flow.bagData.Wait(timeout)
}

func (c *Core) BagSnapshot() []*protocol.ItemSt {
	c.bagMu.RLock()
	defer c.bagMu.RUnlock()
	items := make([]*protocol.ItemSt, 0, len(c.bagItems))
	for _, it := range c.bagItems {
		if it == nil {
			continue
		}
		copyItem := proto.Clone(it).(*protocol.ItemSt)
		items = append(items, copyItem)
	}
	return items
}

func (c *Core) UseItem(itemID, count uint32) error {
	req := &protocol.C2SUseItemReq{
		ItemId: itemID,
		Count:  count,
	}
	return c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SUseItem), req)
}

func (c *Core) WaitUseItemResult(timeout time.Duration) (*protocol.S2CUseItemResultReq, error) {
	return c.flow.useItem.Wait(timeout)
}

func (c *Core) PickupItem(handle uint64) error {
	req := &protocol.C2SPickupItemReq{
		ItemHdl: handle,
	}
	return c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SPickupItem), req)
}

func (c *Core) WaitPickupResult(timeout time.Duration) (*protocol.S2CPickupItemResultReq, error) {
	return c.flow.pickup.Wait(timeout)
}

// --- GM Command ---

func (c *Core) SendGMCommand(name string, args []string) error {
	req := &protocol.C2SGMCommandReq{
		GmName: name,
		Args:   args,
	}
	return c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SGMCommand), req)
}

func (c *Core) WaitGMResult(timeout time.Duration) (*protocol.S2CGMCommandResultReq, error) {
	return c.flow.gmResult.Wait(timeout)
}

// --- Dungeon ---

func (c *Core) EnterDungeonReq(dungeonID, difficulty uint32) error {
	req := &protocol.C2SEnterDungeonReq{
		DungeonId:  dungeonID,
		Difficulty: difficulty,
	}
	return c.sendProtoMessage(uint16(protocol.C2SProtocol_C2SEnterDungeon), req)
}

func (c *Core) WaitEnterDungeonResult(timeout time.Duration) (*protocol.S2CEnterDungeonResultReq, error) {
	return c.flow.dungeonEnter.Wait(timeout)
}

// --- Bag/Money snapshots helpers ---

func (c *Core) MoneySnapshot() map[uint32]int64 {
	c.moneyMu.RLock()
	defer c.moneyMu.RUnlock()
	result := make(map[uint32]int64, len(c.moneyData))
	for k, v := range c.moneyData {
		result[k] = v
	}
	return result
}
