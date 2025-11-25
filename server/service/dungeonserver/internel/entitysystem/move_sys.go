package entitysystem

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/service/dungeonserver/internel/clientprotocol"
	"postapocgame/server/service/dungeonserver/internel/devent"
	"time"

	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/entityhelper"
	"postapocgame/server/service/dungeonserver/internel/gameserverlink"
	"postapocgame/server/service/dungeonserver/internel/iface"

	"google.golang.org/protobuf/proto"
)

// move_sys.go 处理客户端移动协议和位置校验
//
// 移动系统只专注于移动功能，AI相关业务由调用方通过组合调用移动协议实现

const (
	// defaultMaxMoveSpeed 默认最大移动速度（像素/秒）
	defaultMaxMoveSpeed = 600.0
)

// MoveSys 实体移动系统
type MoveSys struct {
	entity iface.IEntity
	scene  iface.IScene

	// 客户端移动状态
	lastTime       int64   // 上一次更新位置的时间（毫秒时间戳），0表示未在移动
	lastX          int32   // 移动开始时的X坐标（像素坐标）
	lastY          int32   // 移动开始时的Y坐标（像素坐标）
	speed          uint32  // 移动速度（像素/秒）
	moveLen        float64 // 移动的总距离（像素）
	lastClientPx   int32   // 上次客户端上报的像素X坐标
	lastClientPy   int32   // 上次客户端上报的像素Y坐标
	lastReportTime time.Time

	moveData *protocol.MoveData // 移动数据，包含目标像素坐标
}

func (ms *MoveSys) logContext() string {
	var sceneID int32
	if ms.scene != nil {
		sceneID = int32(ms.scene.GetSceneId())
	}
	var entityHdl uint64
	if ms.entity != nil {
		entityHdl = ms.entity.GetHdl()
	}
	return fmt.Sprintf("scene=%d entity=%d", sceneID, entityHdl)
}

// NewMoveSys 创建移动系统
func NewMoveSys(entity iface.IEntity) *MoveSys {
	return &MoveSys{
		entity:   entity,
		moveData: &protocol.MoveData{},
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
	ms.StopMove(false, false)
}

// IsMoving 判断是否正在移动
func (ms *MoveSys) IsMoving() bool {
	return ms.lastTime != 0
}

// GetMoveDest 获取移动目标坐标（像素坐标）
func (ms *MoveSys) GetMoveDest() (int32, int32) {
	if ms.moveData == nil {
		return 0, 0
	}
	return ms.moveData.DestPx, ms.moveData.DestPy
}

// ResetState 重置移动状态
func (ms *MoveSys) ResetState() {
	ms.ClearMoveData()
}

// SetLastMoveTime 设置最后移动时间
func (ms *MoveSys) SetLastMoveTime(t int64) {
	ms.lastTime = t
}

// MoveData 获取移动数据
func (ms *MoveSys) MoveData() *protocol.MoveData {
	return ms.moveData
}

// ClearMoveData 清除移动数据
func (ms *MoveSys) ClearMoveData() {
	if ms.moveData != nil {
		ms.moveData.DestPx = -1
		ms.moveData.DestPy = -1
	}
	ms.lastX = 0
	ms.lastY = 0
	ms.lastTime = 0
	ms.moveLen = 0
}

// MovingTime 服务端时间驱动移动
func (ms *MoveSys) MovingTime(mustStop bool) bool {
	if !ms.IsMoving() || ms.scene == nil {
		return false
	}

	et := ms.entity
	if et == nil {
		return false
	}

	if !et.CanBeAttacked() {
		ms.StopMove(true, false)
		return false
	}

	// 获取移动速度（像素/秒）
	speed := ms.speed
	if speed <= 0 {
		speed = uint32(et.GetAttrSys().GetAttrValue(attrdef.AttrMoveSpeed))
		if speed <= 0 {
			speed = uint32(defaultMaxMoveSpeed)
		}
	}

	now := servertime.Now().UnixMilli()
	// 计算从开始移动到现在的时间差（毫秒）
	elapsed := now - ms.lastTime
	if elapsed < 0 {
		elapsed = 0
	}
	// 计算应该移动的像素总距离：速度(像素/秒) * 时间(秒) = 距离(像素)
	totalMoved := float64(int64(speed)*elapsed) / 1000 // 已应移动的总像素距离

	stop := false
	var newX, newY int32

	if ms.moveLen > 0 && totalMoved >= ms.moveLen {
		// 到达目标：已移动距离 >= 总距离，直接设置为目标坐标
		newX, newY = ms.moveData.DestPx, ms.moveData.DestPy
		stop = true
	} else {
		// 未到达目标：按比例计算新位置
		// 计算移动进度比例：已移动距离 / 总距离
		var rate float64
		if ms.moveLen > 0 {
			rate = totalMoved / ms.moveLen
		}
		// 计算目标方向向量（像素坐标差值）
		interX := ms.moveData.DestPx - ms.lastX
		interY := ms.moveData.DestPy - ms.lastY

		// 按比例计算新位置：起始位置 + 方向向量 * 进度比例
		newX = int32(rate*float64(interX)) + ms.lastX
		newY = int32(rate*float64(interY)) + ms.lastY
	}

	// 转换为格子坐标
	gridX := argsdef.PixelCoordToTileX(uint32(newX))
	gridY := argsdef.PixelCoordToTileY(uint32(newY))

	// 检查目标位置是否可行走
	currentPos := et.GetPosition()
	if currentPos == nil {
		return false
	}

	if currentPos.X != gridX || currentPos.Y != gridY {
		if !ms.scene.IsWalkable(int(gridX), int(gridY)) {
			return false
		}
	}

	if stop || mustStop {
		ms.StopMove(mustStop, mustStop)
		return ms.scene.EntityMove(et.GetHdl(), gridX, gridY) == nil
	}

	ms.lastTime = now
	ms.lastX = newX
	ms.lastY = newY
	return ms.scene.EntityMove(et.GetHdl(), gridX, gridY) == nil
}

// HandleStartMove 处理客户端起步移动
func (ms *MoveSys) HandleStartMove(scene iface.IScene, req *protocol.C2SStartMoveReq) error {
	if scene == nil || req == nil {
		return nil
	}

	// 将客户端发送的像素坐标转换为格子坐标
	toTileX, toTileY := argsdef.PixelCoordToTile(req.ToX, req.ToY)
	log.Infof("[MoveSys] %s startMove fromPx=(%d,%d) -> destTile=(%d,%d) speed=%d",
		ms.logContext(), req.FromX, req.FromY, toTileX, toTileY, req.Speed)

	// 处理移动开始
	ret := ms.HandMove(scene, toTileX, toTileY, int32(req.FromX), int32(req.FromY), req.Speed)
	if ret != 0 {
		ms.sendStop()
		return customerr.NewErrorByCode(int32(ret), "start move failed")
	}

	// 广播 S2CStartMove，携带实体hdl和move_data
	ms.BroadcastStartMove(int32(req.FromX), int32(req.FromY))

	return nil
}

// HandMove 处理客户端移动请求
func (ms *MoveSys) HandMove(scene iface.IScene, tileX, tileY uint32, cPx, cPy int32, speed uint32) int32 {
	if scene == nil {
		return int32(protocol.ErrorCode_Param_Invalid)
	}

	// 检查实体是否可以移动
	if !ms.entity.CanBeAttacked() {
		return int32(protocol.ErrorCode_Param_Invalid)
	}

	// 检查目标位置是否可行走
	if !scene.IsWalkable(int(tileX), int(tileY)) {
		log.Warnf("MoveSys:handMove() sceneId:%d, Grid can not move (%d, %d)",
			scene.GetSceneId(), tileX, tileY)
		return int32(protocol.ErrorCode_Param_Invalid)
	}

	// 如果正在移动，先进行位置更新校验（确保客户端位置合理）
	if ms.IsMoving() {
		if !ms.LocationUpdate(cPx, cPy) {
			return int32(protocol.ErrorCode_Param_Invalid)
		}
	}

	// 获取当前位置（格子坐标）
	currentPos := ms.entity.GetPosition()
	if currentPos == nil {
		return int32(protocol.ErrorCode_Internal_Error)
	}

	// 计算目标像素坐标
	destPx, destPy := argsdef.TileCoordToPixel(tileX, tileY)
	tPx := int32(destPx)
	tPy := int32(destPy)

	// 获取当前像素坐标
	currPx, currPy := argsdef.TileCoordToPixel(currentPos.X, currentPos.Y)
	px := int32(currPx)
	py := int32(currPy)

	// 计算移动距离（使用勾股定理：sqrt(dx^2 + dy^2)）
	interX := tPx - px                                           // X方向差值
	interY := tPy - py                                           // Y方向差值
	moveLen := math.Sqrt(float64(interX*interX + interY*interY)) // 总距离（像素）

	// 更新移动数据
	now := servertime.Now().UnixMilli()
	ms.SetLastMoveTime(now)
	ms.moveData.DestPx = tPx
	ms.moveData.DestPy = tPy
	ms.lastX = px
	ms.lastY = py
	ms.lastTime = now
	ms.moveLen = moveLen
	ms.speed = speed
	ms.lastClientPx = cPx
	ms.lastClientPy = cPy
	ms.lastReportTime = servertime.Now()

	log.Infof("[MoveSys] %s planning move startTile=(%d,%d) destTile=(%d,%d) destPx=(%d,%d) distance=%.2fpx speed=%d",
		ms.logContext(), currentPos.X, currentPos.Y, tileX, tileY, tPx, tPy, moveLen, speed)

	return 0
}

// LocationUpdate 位置更新校验
func (ms *MoveSys) LocationUpdate(cPx, cPy int32) bool {
	if !ms.IsMoving() {
		return false
	}

	et := ms.entity
	if et == nil {
		return false
	}

	scene := ms.scene
	if scene == nil {
		return false
	}

	// 获取当前服务端位置（像素坐标）
	currPx, currPy := argsdef.TileCoordToPixel(et.GetPosition().X, et.GetPosition().Y)
	px := int32(currPx)
	py := int32(currPy)

	// 如果客户端位置和服务端位置相同，直接通过
	if cPx == px && cPy == py {
		return true
	}

	// 获取移动速度（像素/秒）
	speed := int32(ms.speed)
	if speed <= 0 {
		speed = int32(et.GetAttrSys().GetAttrValue(attrdef.AttrMoveSpeed))
		if speed <= 0 {
			speed = int32(defaultMaxMoveSpeed)
		}
	}

	// 计算时间差（给与客户端50ping的容忍度）
	milli := servertime.Now().UnixMilli()
	// 计算从上次上报到现在的时间差，加上50ping的容忍度（500ms）
	last := milli - ms.lastTime + 500 // 500ms = 50ping的容忍度
	// 计算允许移动的最大像素距离：速度(像素/秒) * 时间(秒) = 距离(像素)
	pix := int32(float64(speed) * float64(last) / 1000) // 允许移动的最大像素距离

	// 转换为格子坐标进行校验
	gridX := argsdef.PixelCoordToTileX(uint32(cPx))
	gridY := argsdef.PixelCoordToTileY(uint32(cPy))

	// 同步的点不可走，直接拉回
	if !scene.IsWalkable(int(gridX), int(gridY)) {
		log.Warnf("[MoveSys] %s location update rejected: Grid(%d,%d) not walkable (clientPx=%d,%d)",
			ms.logContext(), gridX, gridY, cPx, cPy)
		return false
	}

	// 计算客户端移动距离（像素）：本次位置 - 上次位置
	delPx := cPx - ms.lastClientPx
	delPy := cPy - ms.lastClientPy

	// 校验移动速度：如果移动距离的平方 > 允许距离的平方，说明移动过快
	// 使用平方比较避免开方运算（delPx^2 + delPy^2 > pix^2）
	if delPx*delPx+delPy*delPy > pix*pix {
		log.Warnf("[MoveSys] %s location update too fast last=%dms speed=%d (deltaPx=%d deltaPy=%d) limit=%d dest=(%d,%d) prev=(%d,%d)",
			ms.logContext(), last, speed, delPx, delPy, pix, cPx, cPy, ms.lastClientPx, ms.lastClientPy)
		return false
	}

	// 更新状态
	ms.SetLastMoveTime(milli)
	ms.lastClientPx = cPx
	ms.lastClientPy = cPy

	// 更新服务端位置
	if ms.scene.EntityMove(et.GetHdl(), gridX, gridY) == nil {
		// 检查是否到达目标
		if cPx == ms.moveData.DestPx && cPy == ms.moveData.DestPy {
			ms.StopMove(true, false)
		}
		log.Infof("[MoveSys] %s location update accepted -> tile=(%d,%d) clientPx=(%d,%d)",
			ms.logContext(), gridX, gridY, cPx, cPy)
		return true
	}

	return false
}

// HandleUpdateMove 处理客户端移动更新
func (ms *MoveSys) HandleUpdateMove(scene iface.IScene, req *protocol.C2SUpdateMoveReq) error {
	if scene == nil || req == nil {
		return nil
	}

	// 如果不在移动状态，直接返回
	if !ms.IsMoving() {
		return nil
	}

	// 获取服务端计算的当前位置
	currentPos := ms.entity.GetPosition()
	if currentPos == nil {
		return nil
	}

	// 转换为像素坐标
	serverPxU, serverPyU := argsdef.TileCoordToPixel(currentPos.X, currentPos.Y)
	serverPx := int32(serverPxU)
	serverPy := int32(serverPyU)

	// 客户端上报的坐标
	clientPx := int32(req.PosX)
	clientPy := int32(req.PosY)

	// 计算坐标误差（像素距离）
	dx := float64(clientPx - serverPx)
	dy := float64(clientPy - serverPy)
	distance := math.Sqrt(dx*dx + dy*dy)

	// 获取移动速度（像素/秒）
	speed := float64(ms.speed)
	if speed <= 0 {
		speed = float64(ms.entity.GetAttrSys().GetAttrValue(attrdef.AttrMoveSpeed))
		if speed <= 0 {
			speed = defaultMaxMoveSpeed
		}
	}

	// 支持1s的误差：允许的最大误差距离 = 速度(像素/秒) * 1秒
	maxErrorDistance := speed * 1.0 // 1秒的误差容忍度

	// 如果差距很大，结束移动（防止客户端位置异常）
	if distance > maxErrorDistance {
		log.Warnf("[MoveSys] %s updateMove deviation=%.2f max=%.2f, serverPx=(%d,%d) clientPx=(%d,%d)",
			ms.logContext(), distance, maxErrorDistance, serverPx, serverPy, clientPx, clientPy)
		// 结束移动，最终坐标为服务端当前位置（上一个点）
		// 这样可以防止客户端位置异常导致的瞬移
		ms.HandleEndMove(scene, &protocol.C2SEndMoveReq{
			PosX: uint32(serverPx),
			PosY: uint32(serverPy),
			// Seq 字段已废弃，不再使用
		})
		return nil
	}

	// 否则更新客户端给过来最新的坐标
	// 调用 LocationUpdate 进行位置校验和更新
	if !ms.LocationUpdate(clientPx, clientPy) {
		// 位置校验失败（移动过快、位置不可行走等），结束移动
		// 最终坐标仍为服务端当前位置
		ms.HandleEndMove(scene, &protocol.C2SEndMoveReq{
			PosX: uint32(serverPx),
			PosY: uint32(serverPy),
			// Seq 字段已废弃，不再使用
		})
		return nil
	}

	return nil
}

// HandleEndMove 处理客户端停止移动
func (ms *MoveSys) HandleEndMove(scene iface.IScene, req *protocol.C2SEndMoveReq) error {
	if scene == nil || req == nil {
		return nil
	}

	tileX, tileY := argsdef.PixelCoordToTile(req.PosX, req.PosY)
	log.Infof("[MoveSys] %s endMove reqPx=(%d,%d) -> tile=(%d,%d)", ms.logContext(), req.PosX, req.PosY, tileX, tileY)

	// 停止移动
	ms.StopMove(true, true)

	// 广播 S2CEndMove，通知客户端结束移动
	ms.BroadcastEndMove(int32(req.PosX), int32(req.PosY))

	// 如果是角色实体，同步坐标到GameServer
	ms.syncPositionToGameServer(scene, &argsdef.Position{X: tileX, Y: tileY})

	return nil
}

// BroadcastMove 广播移动
func (ms *MoveSys) BroadcastMove(tileX, tileY, speed uint32) {
	if ms.scene == nil {
		return
	}

	// 转换为像素坐标
	px, py := argsdef.TileCoordToPixel(tileX, tileY)

	resp := &protocol.S2CEntityMoveReq{
		EntityHdl: ms.entity.GetHdl(),
		PosX:      px,
		PosY:      py,
		Speed:     speed,
	}

	data, err := proto.Marshal(resp)
	if err != nil {
		log.Errorf("marshal move payload failed: %v", err)
		return
	}

	// 广播给场景内所有实体
	for _, et := range ms.scene.GetAllEntities() {
		if err := et.SendMessage(uint16(protocol.S2CProtocol_S2CEntityMove), data); err != nil {
			log.Warnf("broadcast move message failed, entity=%d err=%v", et.GetHdl(), err)
		}
	}

	// 也发送给自己
	if err := ms.entity.SendMessage(uint16(protocol.S2CProtocol_S2CEntityMove), data); err != nil {
		log.Warnf("send move message to self failed, entity=%d err=%v", ms.entity.GetHdl(), err)
	}
}

// BroadcastStartMove 广播开始移动
func (ms *MoveSys) BroadcastStartMove(posX, posY int32) {
	if ms.scene == nil || ms.moveData == nil {
		return
	}

	// 构造 MoveData
	moveData := &protocol.MoveData{
		DestPx: ms.moveData.DestPx,
		DestPy: ms.moveData.DestPy,
	}

	resp := &protocol.S2CStartMoveReq{
		EntityHdl: ms.entity.GetHdl(),
		PosX:      uint32(posX),
		PosY:      uint32(posY),
		MoveData:  moveData,
	}

	data, err := proto.Marshal(resp)
	if err != nil {
		log.Errorf("marshal start move payload failed: %v", err)
		return
	}

	// 广播给场景内所有实体
	for _, et := range ms.scene.GetAllEntities() {
		_ = et.SendMessage(uint16(protocol.S2CProtocol_S2CStartMove), data)
	}

	// 也发送给自己
	_ = ms.entity.SendMessage(uint16(protocol.S2CProtocol_S2CStartMove), data)
}

// BroadcastEndMove 广播结束移动
func (ms *MoveSys) BroadcastEndMove(posX, posY int32) {
	if ms.scene == nil {
		return
	}

	resp := &protocol.S2CEndMoveReq{
		EntityHdl: ms.entity.GetHdl(),
		PosX:      uint32(posX),
		PosY:      uint32(posY),
	}

	data, err := proto.Marshal(resp)
	if err != nil {
		log.Errorf("marshal end move payload failed: %v", err)
		return
	}

	// 广播给场景内所有实体
	for _, et := range ms.scene.GetAllEntities() {
		_ = et.SendMessage(uint16(protocol.S2CProtocol_S2CEndMove), data)
	}

	// 也发送给自己
	_ = ms.entity.SendMessage(uint16(protocol.S2CProtocol_S2CEndMove), data)
}

// StopMove 停止移动
func (ms *MoveSys) StopMove(broadcast, sendToSelf bool) {
	if !ms.IsMoving() {
		return
	}

	ms.ClearMoveData()

	// 发送停止移动消息
	pos := ms.entity.GetPosition()
	if pos == nil {
		return
	}

	px, py := argsdef.TileCoordToPixel(pos.X, pos.Y)
	log.Infof("[MoveSys] %s stopMove finalTile=(%d,%d) finalPx=(%d,%d) broadcast=%v self=%v",
		ms.logContext(), pos.X, pos.Y, px, py, broadcast, sendToSelf)
	resp := &protocol.S2CEntityStopMoveReq{
		EntityHdl: ms.entity.GetHdl(),
		PosX:      px,
		PosY:      py,
		// Seq 字段已移除：客户端不使用，参考代码也没有序列号
	}

	data, err := proto.Marshal(resp)
	if err != nil {
		log.Errorf("marshal stop move payload failed: %v", err)
		return
	}

	if sendToSelf {
		_ = ms.entity.SendMessage(uint16(protocol.S2CProtocol_S2CEntityStopMove), data)
	}

	if broadcast && ms.scene != nil {
		for _, et := range ms.scene.GetAllEntities() {
			if et.GetHdl() != ms.entity.GetHdl() {
				_ = et.SendMessage(uint16(protocol.S2CProtocol_S2CEntityStopMove), data)
			}
		}
	}
}

// sendStop 发送停止移动消息给客户端
func (ms *MoveSys) sendStop() {
	pos := ms.entity.GetPosition()
	if pos == nil {
		return
	}
	px, py := argsdef.TileCoordToPixel(pos.X, pos.Y)
	_ = ms.entity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CEntityStopMove), &protocol.S2CEntityStopMoveReq{
		EntityHdl: ms.entity.GetHdl(),
		PosX:      px,
		PosY:      py,
		// Seq 字段已移除：客户端不使用，参考代码也没有序列号
	})
}

// flushAOIChanges 刷新AOI变化
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
	_ = receiver.SendProtoMessage(uint16(protocol.S2CProtocol_S2CEntityAppear), &protocol.S2CEntityAppearReq{
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
	_ = receiver.SendProtoMessage(uint16(protocol.S2CProtocol_S2CEntityDisappear), &protocol.S2CEntityDisappearReq{
		EntityHdl: targetHdl,
	})
}

func (ms *MoveSys) notifyDisappearByHandle(receiverHdl uint64, subjectHdl uint64) {
	if ms.scene == nil {
		return
	}
	target, ok := ms.scene.GetEntity(receiverHdl)
	if !ok || target.GetEntityType() != uint32(protocol.EntityType_EtRole) {
		return
	}
	_ = target.SendProtoMessage(uint16(protocol.S2CProtocol_S2CEntityDisappear), &protocol.S2CEntityDisappearReq{
		EntityHdl: subjectHdl,
	})
}

// syncPositionToGameServer 同步坐标到GameServer
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
	reqData, err := proto.Marshal(&protocol.D2GSyncPositionReq{
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

// 协议处理函数
func handleStartMove(entity iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SStartMoveReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	scene, err := getSceneByEntity(entity)
	if err != nil {
		return err
	}
	moveSys := entity.GetMoveSys()
	if moveSys == nil {
		return nil
	}
	return moveSys.HandleStartMove(scene, &req)
}

func handleUpdateMove(entity iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SUpdateMoveReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	scene, err := getSceneByEntity(entity)
	if err != nil {
		return err
	}
	moveSys := entity.GetMoveSys()
	if moveSys == nil {
		return nil
	}
	return moveSys.HandleUpdateMove(scene, &req)
}

func handleEndMove(entity iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SEndMoveReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	scene, err := getSceneByEntity(entity)
	if err != nil {
		return err
	}
	moveSys := entity.GetMoveSys()
	if moveSys == nil {
		return nil
	}
	return moveSys.HandleEndMove(scene, &req)
}

func handleGetNearestMonster(role iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SGetNearestMonsterReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	scene, err := getSceneByEntity(role)
	if err != nil {
		return err
	}

	// 获取角色位置
	rolePos := role.GetPosition()
	if rolePos == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "role position not found")
	}

	// 获取场景中所有实体
	allEntities := scene.GetAllEntities()

	// 查找最近的怪物
	var nearestMonster iface.IEntity
	var minDistance float64 = math.MaxFloat64

	for _, e := range allEntities {
		// 只处理怪物实体
		if e.GetEntityType() != uint32(protocol.EntityType_EtMonster) {
			continue
		}

		// 跳过死亡的怪物
		if e.IsDead() {
			continue
		}

		// 计算距离
		monsterPos := e.GetPosition()
		if monsterPos == nil {
			continue
		}

		dx := float64(monsterPos.X) - float64(rolePos.X)
		dy := float64(monsterPos.Y) - float64(rolePos.Y)
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance < minDistance {
			minDistance = distance
			nearestMonster = e
		}
	}

	// 构造响应
	resp := &protocol.S2CGetNearestMonsterResultReq{
		Success: nearestMonster != nil,
	}

	if nearestMonster != nil {
		monsterPos := nearestMonster.GetPosition()
		monsterEntity, ok := nearestMonster.(iface.IMonster)
		if ok {
			resp.MonsterHdl = nearestMonster.GetHdl()
			resp.MonsterId = uint32(monsterEntity.GetId())
			resp.MonsterName = monsterEntity.GetName()
			if monsterPos != nil {
				resp.PosX = monsterPos.X
				resp.PosY = monsterPos.Y
			}
			resp.Distance = float32(minDistance)
			resp.Message = "找到最近怪物"
		} else {
			resp.Success = false
			resp.Message = "怪物实体类型错误"
		}
	} else {
		resp.Message = "未找到怪物"
	}

	return role.SendProtoMessage(uint16(protocol.S2CProtocol_S2CGetNearestMonsterResult), resp)
}

func handleChangeScene(entity iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SChangeSceneReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	scene, err := getSceneByEntity(entity)
	if err != nil {
		return err
	}

	// 获取当前场景ID和副本ID
	currentSceneId := entity.GetSceneId()
	currentFuBenId := scene.GetFuBenId()
	targetSceneId := req.SceneId

	// 获取当前副本实例
	fuBen := scene.GetFuBen()
	if fuBen == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "副本不存在")
	}

	// 获取目标场景
	targetScene := fuBen.GetScene(targetSceneId)
	if targetScene == nil {
		resp := &protocol.S2CChangeSceneResultReq{
			Success: false,
			Message: "目标场景不存在",
			SceneId: targetSceneId,
		}
		return entity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CChangeSceneResult), resp)
	}

	// 检查是否在同一副本内（不同副本不能跨副本场景切换）
	if targetScene.GetFuBenId() != currentFuBenId {
		resp := &protocol.S2CChangeSceneResultReq{
			Success: false,
			Message: "不能跨副本切换场景",
			SceneId: targetSceneId,
		}
		return entity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CChangeSceneResult), resp)
	}

	// 如果目标场景就是当前场景，直接返回成功
	if targetSceneId == currentSceneId {
		resp := &protocol.S2CChangeSceneResultReq{
			Success: true,
			Message: "切换成功",
			SceneId: targetSceneId,
		}
		return entity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CChangeSceneResult), resp)
	}

	// 从当前场景移除实体
	err = scene.RemoveEntity(entity.GetHdl())
	if err != nil {
		log.Errorf("scene RemoveEntity %s %d failed, err:%v", entity.GetName(), entity.GetId(), err)
	}

	// 将实体添加到目标场景
	// 从场景配置获取出生点
	configMgr := jsonconf.GetConfigManager()
	sceneConfig, _ := configMgr.GetSceneConfig(targetSceneId)
	var x, y uint32
	if sceneConfig != nil && sceneConfig.BornArea != nil {
		// 从出生点范围随机选择
		bornArea := sceneConfig.BornArea
		if bornArea.X2 > bornArea.X1 && bornArea.Y2 > bornArea.Y1 {
			x = bornArea.X1 + uint32(rand.Intn(int(bornArea.X2-bornArea.X1)))
			y = bornArea.Y1 + uint32(rand.Intn(int(bornArea.Y2-bornArea.Y1)))
		} else {
			// 使用默认位置
			x, y = 100, 100
		}
	} else {
		// 使用默认位置
		x, y = 100, 100
	}
	entity.SetPosition(x, y)
	err = targetScene.AddEntity(entity)
	if err != nil {
		log.Errorf("scene AddEntity %s %d failed, err:%v", entity.GetName(), entity.GetId(), err)
	}

	log.Infof("Entity %d changed scene from %d to %d", entity.GetHdl(), currentSceneId, targetSceneId)

	// 发送切换成功响应
	resp := &protocol.S2CChangeSceneResultReq{
		Success: true,
		Message: "切换成功",
		SceneId: targetSceneId,
	}
	return entity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CChangeSceneResult), resp)
}

func init() {
	devent.Subscribe(devent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SStartMove), handleStartMove)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SUpdateMove), handleUpdateMove)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SEndMove), handleEndMove)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SGetNearestMonster), handleGetNearestMonster)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SChangeScene), handleChangeScene)
	})
}
