package entitysystem

import (
	"context"
	"math"
	"sync"
	"time"

	"postapocgame/server/internal"
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/entityhelper"
	"postapocgame/server/service/dungeonserver/internel/gameserverlink"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

const (
	autoMoveMinInterval     = 80 * time.Millisecond
	autoMoveTolerance       = 2.0
	defaultMaxMoveSpeed     = 600.0
	startMoveWindowDuration = time.Second
	minDeltaDuration        = 50 * time.Millisecond
	positionTolerance       = 30.0
	distanceTolerance       = 50.0
)

// MoveState 移动状态
type MoveState struct {
	LastSeq        uint32
	LastReportTime time.Time
	LastClientPos  argsdef.Position
	IsMoving       bool
}

// MoveSys 实体移动系统（兼容玩家与AI）
type MoveSys struct {
	entity iface.IEntity
	scene  iface.IScene

	// 自动移动（AI用）
	autoTarget *argsdef.Position
	autoSpeed  float64
	autoMoving bool
	autoSeq    uint32
	lastTick   time.Time

	// 客户端移动状态（玩家用）
	mu    sync.Mutex
	state *MoveState
}

// NewMoveSys 创建移动系统
func NewMoveSys(entity iface.IEntity) *MoveSys {
	return &MoveSys{
		entity: entity,
	}
}

// BindScene 绑定场景
func (ms *MoveSys) BindScene(scene iface.IScene) {
	ms.scene = scene
}

// UnbindScene 解绑场景
func (ms *MoveSys) UnbindScene(scene iface.IScene) {
	if ms.scene == scene {
		ms.scene = nil
	}
	ms.Stop()
}

// RunOne 自动移动驱动（用于AI）
func (ms *MoveSys) RunOne(now time.Time) {
	if !ms.autoMoving || ms.scene == nil || ms.autoTarget == nil {
		return
	}
	if !ms.lastTick.IsZero() && now.Sub(ms.lastTick) < autoMoveMinInterval {
		return
	}
	ms.lastTick = now

	current := ms.entity.GetPosition()
	if current == nil {
		return
	}

	dx := float64(int64(ms.autoTarget.X) - int64(current.X))
	dy := float64(int64(ms.autoTarget.Y) - int64(current.Y))
	dist := math.Hypot(dx, dy)
	if dist <= autoMoveTolerance {
		ms.completeAutoMove()
		return
	}

	speed := ms.autoSpeed
	if speed <= 0 {
		speed = float64(ms.entity.GetAttrSys().GetAttrValue(attrdef.AttrMoveSpeed))
		if speed <= 0 {
			speed = 100
		}
	}

	sec := autoMoveMinInterval.Seconds()
	step := speed * sec
	if step <= 0 {
		step = 1
	}
	ratio := step / dist
	if ratio > 1 {
		ratio = 1
	}

	nx := float64(current.X) + dx*ratio
	ny := float64(current.Y) + dy*ratio

	newPos := &argsdef.Position{
		X: clampToUint32(nx),
		Y: clampToUint32(ny),
	}

	if err := ms.scene.EntityMove(ms.entity.GetHdl(), newPos.X, newPos.Y); err != nil {
		log.Warnf("auto move failed: %v", err)
		ms.Stop()
		return
	}

	ms.autoSeq++
	ms.broadcastMove(uint16(protocol.S2CProtocol_S2CEntityMove), newPos.X, newPos.Y, uint32(speed), ms.autoSeq)
	ms.flushAOIChanges()

	if distanceBetween(newPos, ms.autoTarget) <= autoMoveTolerance {
		ms.completeAutoMove()
	}
}

// MoveTo 设置自动移动目标
func (ms *MoveSys) MoveTo(pos *argsdef.Position, speed float64) {
	if pos == nil {
		return
	}
	if ms.autoTarget != nil && ms.autoTarget.X == pos.X && ms.autoTarget.Y == pos.Y {
		if speed > 0 {
			ms.autoSpeed = speed
		}
		ms.autoMoving = true
		return
	}
	ms.autoTarget = &argsdef.Position{X: pos.X, Y: pos.Y}
	ms.autoSpeed = speed
	ms.autoMoving = true
	ms.lastTick = time.Time{}
}

// Stop 停止自动移动
func (ms *MoveSys) Stop() {
	ms.autoMoving = false
	ms.autoTarget = nil
}

func (ms *MoveSys) completeAutoMove() {
	if ms.autoTarget != nil {
		ms.broadcastMove(uint16(protocol.S2CProtocol_S2CEntityStopMove), ms.autoTarget.X, ms.autoTarget.Y, 0, ms.autoSeq)
		ms.flushAOIChanges()
	}
	ms.autoMoving = false
}

// ResetState 重置移动状态
func (ms *MoveSys) ResetState() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.ensureStateLocked(true)
}

// HandleStartMove 处理客户端起步移动
func (ms *MoveSys) HandleStartMove(scene iface.IScene, req *protocol.C2SStartMoveReq) error {
	if scene == nil || req == nil {
		return nil
	}
	pos, speed, err := ms.handleStart(scene, req.Seq, req.FromX, req.FromY, req.ToX, req.ToY, req.Speed)
	if err != nil {
		ms.sendStop(req.Seq)
		return err
	}
	if err := ms.applySceneMove(scene, pos); err != nil {
		ms.sendStop(req.Seq)
		return err
	}
	ms.broadcastMove(uint16(protocol.S2CProtocol_S2CEntityMove), pos.X, pos.Y, speed, req.Seq)
	ms.flushAOIChanges()
	return nil
}

// HandleUpdateMove 处理客户端移动更新
func (ms *MoveSys) HandleUpdateMove(scene iface.IScene, req *protocol.C2SUpdateMoveReq) error {
	if scene == nil || req == nil {
		return nil
	}
	pos, speed, err := ms.handleUpdate(scene, req.Seq, req.PosX, req.PosY, req.Speed)
	if err != nil {
		ms.sendStop(req.Seq)
		return err
	}
	if err := ms.applySceneMove(scene, pos); err != nil {
		ms.sendStop(req.Seq)
		return err
	}
	ms.broadcastMove(uint16(protocol.S2CProtocol_S2CEntityMove), pos.X, pos.Y, speed, req.Seq)
	ms.flushAOIChanges()
	return nil
}

// HandleEndMove 处理客户端停止移动
func (ms *MoveSys) HandleEndMove(scene iface.IScene, req *protocol.C2SEndMoveReq) error {
	if scene == nil || req == nil {
		return nil
	}
	pos, err := ms.handleEnd(scene, req.Seq, req.PosX, req.PosY)
	if err != nil {
		ms.sendStop(req.Seq)
		return err
	}
	if err := ms.applySceneMove(scene, pos); err != nil {
		ms.sendStop(req.Seq)
		return err
	}
	ms.broadcastMove(uint16(protocol.S2CProtocol_S2CEntityStopMove), pos.X, pos.Y, 0, req.Seq)
	ms.flushAOIChanges()

	// 如果是角色实体，同步坐标到GameServer
	ms.syncPositionToGameServer(scene, pos)

	return nil
}

// syncPositionToGameServer 同步坐标到GameServer（仅对角色实体）
func (ms *MoveSys) syncPositionToGameServer(scene iface.IScene, pos *argsdef.Position) {
	if ms.entity == nil || scene == nil || pos == nil {
		return
	}

	// 只处理角色实体
	if ms.entity.GetEntityType() != uint32(protocol.EntityType_EtRole) {
		return
	}

	// 类型断言为RoleEntity
	roleEntity, ok := ms.entity.(interface {
		GetSessionId() string
		GetRoleId() uint64
	})
	if !ok {
		return
	}

	sessionId := roleEntity.GetSessionId()
	roleId := roleEntity.GetRoleId()
	if sessionId == "" || roleId == 0 {
		return
	}

	// 构造RPC请求
	reqData, err := internal.Marshal(&protocol.D2GSyncPositionReq{
		SessionId: sessionId,
		RoleId:    roleId,
		SceneId:   scene.GetSceneId(),
		PosX:      pos.X,
		PosY:      pos.Y,
	})
	if err != nil {
		log.Errorf("marshal sync position request failed: %v", err)
		return
	}

	// 异步调用GameServer（不等待响应）
	err = gameserverlink.CallGameServer(context.Background(), sessionId, uint16(protocol.D2GRpcProtocol_D2GSyncPosition), reqData)
	if err != nil {
		log.Errorf("call gameserver sync position failed: %v", err)
		// 不返回错误，坐标同步失败不影响移动
	}
}

func (ms *MoveSys) applySceneMove(scene iface.IScene, pos *argsdef.Position) error {
	if scene == nil || pos == nil {
		return nil
	}
	if err := scene.EntityMove(ms.entity.GetHdl(), pos.X, pos.Y); err != nil {
		return err
	}
	return nil
}

func (ms *MoveSys) broadcastMove(protoId uint16, x, y, speed, seq uint32) {
	if ms.scene == nil {
		return
	}
	resp := &protocol.S2CEntityMoveReq{
		EntityHdl: ms.entity.GetHdl(),
		PosX:      x,
		PosY:      y,
		Speed:     speed,
		Seq:       seq,
	}
	if protoId == uint16(protocol.S2CProtocol_S2CEntityStopMove) {
		resp = nil
	}

	var payload interface{}
	if protoId == uint16(protocol.S2CProtocol_S2CEntityMove) {
		payload = resp
	} else {
		payload = &protocol.S2CEntityStopMoveReq{
			EntityHdl: ms.entity.GetHdl(),
			PosX:      x,
			PosY:      y,
			Seq:       seq,
		}
	}

	data, err := internal.Marshal(payload)
	if err != nil {
		log.Errorf("marshal move payload failed: %v", err)
		return
	}
	for _, et := range ms.scene.GetAllEntities() {
		_ = et.SendMessage(protoId, data)
	}
}

func (ms *MoveSys) sendStop(seq uint32) {
	pos := ms.entity.GetPosition()
	if pos == nil {
		return
	}
	_ = ms.entity.SendJsonMessage(uint16(protocol.S2CProtocol_S2CEntityStopMove), &protocol.S2CEntityStopMoveReq{
		EntityHdl: ms.entity.GetHdl(),
		PosX:      pos.X,
		PosY:      pos.Y,
		Seq:       seq,
	})
}

func (ms *MoveSys) flushAOIChanges() {
	if ms.entity == nil {
		return
	}
	aoi := ms.entity.GetAOISys()
	if aoi == nil {
		return
	}
	enterEntities, leaveHandles := aoi.ConsumeVisibilityChanges()
	if len(enterEntities) > 0 {
		for _, target := range enterEntities {
			ms.notifyAppear(ms.entity, target)
			ms.notifyAppear(target, ms.entity)
		}
	}
	if len(leaveHandles) > 0 {
		for _, hdl := range leaveHandles {
			ms.notifyDisappear(ms.entity, hdl)
			ms.notifyDisappearByHandle(hdl, ms.entity.GetHdl())
		}
	}
}

func (ms *MoveSys) notifyAppear(receiver iface.IEntity, subject iface.IEntity) {
	if receiver == nil || subject == nil {
		return
	}
	if receiver.GetEntityType() != uint32(protocol.EntityType_EtRole) {
		return
	}
	entitySt := entityhelper.BuildEntitySnapshot(subject)
	if entitySt == nil {
		return
	}
	_ = receiver.SendJsonMessage(uint16(protocol.S2CProtocol_S2CEntityAppear), &protocol.S2CEntityAppearReq{
		Entity: entitySt,
	})
}

func (ms *MoveSys) notifyDisappear(receiver iface.IEntity, targetHdl uint64) {
	if receiver == nil {
		return
	}
	if receiver.GetEntityType() != uint32(protocol.EntityType_EtRole) {
		return
	}
	_ = receiver.SendJsonMessage(uint16(protocol.S2CProtocol_S2CEntityDisappear), &protocol.S2CEntityDisappearReq{
		EntityHdl: targetHdl,
	})
}

func (ms *MoveSys) notifyDisappearByHandle(receiverHdl uint64, subjectHdl uint64) {
	// 用于在对方是玩家时通知其看不到当前实体
	if ms.scene == nil {
		return
	}
	target, ok := ms.scene.GetEntity(receiverHdl)
	if !ok || target.GetEntityType() != uint32(protocol.EntityType_EtRole) {
		return
	}
	_ = target.SendJsonMessage(uint16(protocol.S2CProtocol_S2CEntityDisappear), &protocol.S2CEntityDisappearReq{
		EntityHdl: subjectHdl,
	})
}

// handleStart 处理移动开始
func (ms *MoveSys) handleStart(scene iface.IScene, seq uint32, fromX, fromY, toX, toY, speed uint32) (*argsdef.Position, uint32, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	state := ms.ensureStateLocked(false)
	if state == nil {
		return nil, 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "movement state not found")
	}

	if scene == nil {
		return nil, 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "scene not found")
	}

	if seq == 0 || (state.LastSeq != 0 && seq <= state.LastSeq) {
		return nil, 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "seq invalid")
	}

	speedLimit := ms.getSpeedLimit()
	if speed == 0 || float64(speed) > speedLimit {
		return nil, 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "speed invalid")
	}

	fromPos := argsdef.Position{X: fromX, Y: fromY}
	toPos := argsdef.Position{X: toX, Y: toY}

	if distanceBetween(ms.entity.GetPosition(), &fromPos) > positionTolerance {
		return nil, 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "position desync")
	}

	if !scene.IsWalkable(int(toPos.X), int(toPos.Y)) {
		return nil, 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "target not walkable")
	}

	travelDistance := distanceBetween(&fromPos, &toPos)
	maxAllowed := speedLimit*startMoveWindowDuration.Seconds() + distanceTolerance
	if travelDistance > maxAllowed {
		return nil, 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "movement too fast")
	}

	state.LastSeq = seq
	state.LastClientPos = toPos
	state.LastReportTime = time.Now()
	state.IsMoving = true

	return &toPos, speed, nil
}

// handleUpdate 处理移动更新
func (ms *MoveSys) handleUpdate(scene iface.IScene, seq, posX, posY, speed uint32) (*argsdef.Position, uint32, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	state := ms.ensureStateLocked(false)
	if state == nil {
		return nil, 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "movement state not found")
	}

	if !state.IsMoving {
		return nil, 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "not moving")
	}

	if seq == 0 || seq <= state.LastSeq {
		return nil, 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "seq invalid")
	}

	if scene == nil {
		return nil, 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "scene not found")
	}

	speedLimit := ms.getSpeedLimit()
	if speed == 0 || float64(speed) > speedLimit {
		return nil, 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "speed invalid")
	}

	targetPos := argsdef.Position{X: posX, Y: posY}

	if !scene.IsWalkable(int(targetPos.X), int(targetPos.Y)) {
		return nil, 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "target not walkable")
	}

	delta := time.Since(state.LastReportTime)
	if delta < minDeltaDuration {
		delta = minDeltaDuration
	}

	travelDistance := distanceBetween(&state.LastClientPos, &targetPos)
	maxAllowed := speedLimit*delta.Seconds() + distanceTolerance
	if travelDistance > maxAllowed {
		return nil, 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "movement too fast")
	}

	state.LastSeq = seq
	state.LastClientPos = targetPos
	state.LastReportTime = time.Now()

	return &targetPos, speed, nil
}

// handleEnd 处理移动结束
func (ms *MoveSys) handleEnd(scene iface.IScene, seq, posX, posY uint32) (*argsdef.Position, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	state := ms.ensureStateLocked(false)
	if state == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "movement state not found")
	}

	if seq == 0 || seq < state.LastSeq {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "seq invalid")
	}

	targetPos := argsdef.Position{X: posX, Y: posY}
	if scene == nil || !scene.IsWalkable(int(targetPos.X), int(targetPos.Y)) {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "target not walkable")
	}

	travelDistance := distanceBetween(&state.LastClientPos, &targetPos)
	if travelDistance > distanceTolerance*2 {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "teleport detected")
	}

	state.LastSeq = seq
	state.LastClientPos = targetPos
	state.LastReportTime = time.Now()
	state.IsMoving = false

	return &targetPos, nil
}

// ensureStateLocked 确保状态存在（必须在锁内调用）
func (ms *MoveSys) ensureStateLocked(force bool) *MoveState {
	if ms.entity == nil {
		return nil
	}
	if ms.state == nil || force {
		pos := ms.entity.GetPosition()
		if pos == nil {
			return nil
		}
		ms.state = &MoveState{
			LastClientPos:  *pos,
			LastReportTime: time.Now(),
		}
	}
	return ms.state
}

// getSpeedLimit 获取速度限制
func (ms *MoveSys) getSpeedLimit() float64 {
	if ms.entity == nil {
		return defaultMaxMoveSpeed
	}
	val := ms.entity.GetAttrSys().GetAttrValue(attrdef.AttrMoveSpeed)
	if val <= 0 {
		return defaultMaxMoveSpeed
	}
	return float64(val)
}
