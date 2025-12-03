// ai_sys.go 定义怪物 AI 系统，负责巡逻、追击、施法等状态切换。
package entitysystem

import (
	"math"
	"math/rand"
	"time"

	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entitymgr"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/iface"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/skill"
)

const (
	DefaultMonsterSkillID uint32        = 1001
	minStepDistance       float64       = 6.0
	aiMoveUpdateInterval  time.Duration = 200 * time.Millisecond // AI移动更新间隔（模拟客户端逐格上报）
)

// AISys 怪物AI系统
type AISys struct {
	entity           iface.IEntity
	scene            iface.IScene
	aiConfig         jsonconf.MonsterAIConfig
	state            AIState
	target           iface.IEntity
	homePos          argsdef.Position
	patrolTarget     *argsdef.Position
	rand             *rand.Rand
	lastAttack       time.Time
	preferredSkill   uint32
	thinkInterval    time.Duration
	nextThink        time.Time
	pathfindingPath  []*argsdef.Position // 当前寻路路径
	pathfindingIndex int                 // 当前路径索引

	// AI移动状态（模拟客户端移动）
	aiMoveTarget     *argsdef.Position // 当前移动目标
	aiMoveStartTime  time.Time         // 移动开始时间
	aiMoveLastUpdate time.Time         // 上次移动更新时间
}

// NewAISys 创建AI系统
func NewAISys(entity iface.IEntity, monsterCfg *jsonconf.MonsterConfig) *AISys {
	if entity == nil || monsterCfg == nil {
		return nil
	}
	cfg := monsterCfg.AIConfig.WithDefaults()
	stepRand := rand.New(rand.NewSource(servertime.Now().UnixNano()))

	preferredSkill := DefaultMonsterSkillID
	if len(monsterCfg.SkillIds) > 0 {
		preferredSkill = monsterCfg.SkillIds[0]
	}

	interval := time.Duration(cfg.ThinkIntervalMS) * time.Millisecond
	if interval < 200*time.Millisecond {
		interval = 200 * time.Millisecond
	}

	return &AISys{
		entity:         entity,
		aiConfig:       cfg,
		state:          AIStateIdle,
		rand:           stepRand,
		preferredSkill: preferredSkill,
		thinkInterval:  interval,
	}
}

func (ai *AISys) SetHomePosition(pos *argsdef.Position) {
	if pos == nil {
		return
	}
	ai.homePos = *pos
}

// RunOne 由Actor主循环驱动
func (ai *AISys) RunOne(now time.Time) {
	if ai == nil || ai.entity == nil || ai.entity.IsDead() {
		return
	}

	// 驱动移动更新（模拟客户端逐格上报）
	ai.tickMoveUpdate(now)

	if ai.nextThink.After(now) {
		return
	}
	ai.nextThink = now.Add(ai.thinkInterval)

	if ai.scene == nil {
		if scene, ok := entitymgr.GetEntityMgr().GetSceneByHandle(ai.entity.GetHdl()); ok {
			ai.scene = scene
		} else {
			return
		}
	}

	if ai.homePos == (argsdef.Position{}) {
		if pos := ai.entity.GetPosition(); pos != nil {
			ai.homePos = *pos
		}
	}

	ai.refreshTarget()

	switch ai.state {
	case AIStateIdle, AIStatePatrol:
		ai.handleIdle()
	case AIStateChase:
		ai.handleChase()
	case AIStateAttack:
		ai.handleAttack()
	case AIStateReturning:
		ai.handleReturn()
	}
}

// NotifyDie 怪物死亡
func (ai *AISys) NotifyDie() {
	ai.target = nil
	ai.state = AIStateIdle
	ai.aiMoveTarget = nil
}

func (ai *AISys) refreshTarget() {
	if ai.target != nil {
		if ai.target.IsDead() {
			ai.target = nil
		} else {
			dist := ai.distanceTo(ai.target.GetPosition())
			if dist > float64(ai.aiConfig.ResetDistance) {
				ai.target = nil
				ai.state = AIStateReturning
			}
		}
	}

	if ai.target == nil {
		if target := ai.findNearestPlayer(); target != nil {
			ai.target = target
			ai.state = AIStateChase
		}
	}

	if ai.target == nil && ai.state == AIStateChase {
		ai.state = AIStateReturning
	}
}

func (ai *AISys) handleIdle() {
	if ai.target != nil {
		ai.state = AIStateChase
		return
	}
	if ai.patrolTarget == nil || ai.distanceTo(ai.patrolTarget) < 4 {
		ai.patrolTarget = ai.randomPatrolPoint()
		ai.pathfindingPath = nil // 重置寻路路径
		ai.pathfindingIndex = 0
	}
	if ai.patrolTarget != nil {
		if ai.moveTowardsWithPathfinding(ai.patrolTarget, ai.aiConfig.PatrolPathfinding) {
			ai.patrolTarget = nil
			ai.pathfindingPath = nil
			ai.pathfindingIndex = 0
		}
		ai.state = AIStatePatrol
	}
}

func (ai *AISys) handleChase() {
	if ai.target == nil {
		ai.state = AIStateReturning
		ai.pathfindingPath = nil
		ai.pathfindingIndex = 0
		return
	}
	dist := ai.distanceTo(ai.target.GetPosition())
	if dist <= float64(ai.aiConfig.AttackRange) {
		ai.state = AIStateAttack
		ai.pathfindingPath = nil
		ai.pathfindingIndex = 0
		return
	}
	// 追击时使用配置的寻路算法
	targetPos := ai.target.GetPosition()
	if targetPos != nil {
		ai.moveTowardsWithPathfinding(targetPos, ai.aiConfig.ChasePathfinding)
	}
}

func (ai *AISys) handleAttack() {
	if ai.target == nil {
		ai.state = AIStateReturning
		return
	}
	dist := ai.distanceTo(ai.target.GetPosition())
	if dist > float64(ai.aiConfig.AttackRange)+20 {
		ai.state = AIStateChase
		return
	}
	if time.Since(ai.lastAttack) < 600*time.Millisecond {
		return
	}
	if ai.castSkill(ai.target) {
		ai.lastAttack = servertime.Now()
	}
}

func (ai *AISys) handleReturn() {
	dist := ai.distanceTo(&ai.homePos)
	if dist <= 5 {
		ai.state = AIStateIdle
		ai.target = nil
		ai.pathfindingPath = nil
		ai.pathfindingIndex = 0
		return
	}
	// 返回时使用A*寻路（确保能绕过障碍物）
	ai.moveTowardsWithPathfinding(&ai.homePos, jsonconf.PathfindingTypeAStar)
}

func (ai *AISys) findNearestPlayer() iface.IEntity {
	if ai.scene == nil {
		return nil
	}
	selfPos := ai.entity.GetPosition()
	var selected iface.IEntity
	minDist := math.MaxFloat64

	for _, et := range ai.scene.GetAllEntities() {
		if et == nil || et.IsDead() {
			continue
		}
		if et.GetEntityType() != uint32(protocol.EntityType_EtRole) {
			continue
		}
		dist := distanceBetween(selfPos, et.GetPosition())
		if dist <= float64(ai.aiConfig.DetectRange) && dist < minDist {
			minDist = dist
			selected = et
		}
	}
	return selected
}

func (ai *AISys) moveTowards(target *argsdef.Position) bool {
	if target == nil || ai.scene == nil {
		return false
	}
	moveSys := ai.entity.GetMoveSys()
	if moveSys == nil {
		return false
	}

	current := ai.entity.GetPosition()
	if current == nil {
		return false
	}

	// 如果已经在目标点附近，完成移动
	dist := ai.distanceTo(target)
	if dist < 3 {
		return true
	}

	// 如果正在移动，先结束当前移动
	if moveSys.IsMoving() {
		currentPx, currentPy := argsdef.TileCoordToPixel(current.X, current.Y)
		moveSys.HandleEndMove(ai.scene, &protocol.C2SEndMoveReq{
			PosX: uint32(currentPx),
			PosY: uint32(currentPy),
		})
	}

	// 像客户端那样调用 HandleStartMove
	fromPx, fromPy := argsdef.TileCoordToPixel(current.X, current.Y)
	toPx, toPy := argsdef.TileCoordToPixel(target.X, target.Y)
	speed := uint32(ai.entity.GetAttrSys().GetAttrValue(attrdef.AttrMoveSpeed))
	if speed <= 0 {
		speed = uint32(ai.moveStep())
	}

	err := moveSys.HandleStartMove(ai.scene, &protocol.C2SStartMoveReq{
		FromX: uint32(fromPx),
		FromY: uint32(fromPy),
		ToX:   uint32(toPx),
		ToY:   uint32(toPy),
		Speed: speed,
	})
	if err != nil {
		return false
	}

	// 记录移动目标，用于后续更新
	ai.aiMoveTarget = target
	ai.aiMoveStartTime = servertime.Now()
	ai.aiMoveLastUpdate = servertime.Now()

	return false
}

// moveTowardsWithPathfinding 使用寻路算法移动到目标点
func (ai *AISys) moveTowardsWithPathfinding(target *argsdef.Position, pathType jsonconf.PathfindingType) bool {
	if target == nil || ai.scene == nil {
		return false
	}
	moveSys := ai.entity.GetMoveSys()
	if moveSys == nil {
		return false
	}

	current := ai.entity.GetPosition()
	if current == nil {
		return false
	}

	// 如果已经在目标点附近，完成移动
	dist := ai.distanceTo(target)
	if dist < 3 {
		return true
	}

	// 检查是否需要重新寻路
	needRepath := false
	if len(ai.pathfindingPath) == 0 || ai.pathfindingIndex >= len(ai.pathfindingPath) {
		needRepath = true
	} else {
		// 检查当前目标点是否仍然有效
		currentTarget := ai.pathfindingPath[ai.pathfindingIndex]
		distToCurrentTarget := distanceBetween(current, currentTarget)
		if distToCurrentTarget < 2 {
			// 到达当前路径点，移动到下一个
			ai.pathfindingIndex++
			if ai.pathfindingIndex >= len(ai.pathfindingPath) {
				needRepath = true
			}
		}
	}

	// 重新寻路
	if needRepath {
		pathResult := FindPath(ai.scene, current.X, current.Y, target.X, target.Y, pathType)
		if !pathResult.Found || len(pathResult.Path) == 0 {
			// 寻路失败，使用直接移动
			ai.pathfindingPath = nil
			ai.pathfindingIndex = 0
			return ai.moveTowards(target)
		}
		ai.pathfindingPath = pathResult.Path
		ai.pathfindingIndex = 0
	}

	// 移动到当前路径点
	if ai.pathfindingIndex < len(ai.pathfindingPath) {
		nextPos := ai.pathfindingPath[ai.pathfindingIndex]
		return ai.moveTowards(nextPos)
	}

	return false
}

// tickMoveUpdate 驱动移动更新（模拟客户端逐格上报）
func (ai *AISys) tickMoveUpdate(now time.Time) {
	if ai.scene == nil {
		return
	}
	moveSys := ai.entity.GetMoveSys()
	if moveSys == nil {
		return
	}

	// 如果不在移动状态，直接返回
	if !moveSys.IsMoving() {
		ai.aiMoveTarget = nil
		return
	}

	// 如果没有移动目标，结束移动
	if ai.aiMoveTarget == nil {
		current := ai.entity.GetPosition()
		if current != nil {
			currentPx, currentPy := argsdef.TileCoordToPixel(current.X, current.Y)
			moveSys.HandleEndMove(ai.scene, &protocol.C2SEndMoveReq{
				PosX: uint32(currentPx),
				PosY: uint32(currentPy),
			})
		}
		return
	}

	// 检查是否到达目标
	current := ai.entity.GetPosition()
	if current != nil {
		dist := ai.distanceTo(ai.aiMoveTarget)
		if dist < 2 {
			// 到达目标，结束移动
			targetPx, targetPy := argsdef.TileCoordToPixel(ai.aiMoveTarget.X, ai.aiMoveTarget.Y)
			moveSys.HandleEndMove(ai.scene, &protocol.C2SEndMoveReq{
				PosX: uint32(targetPx),
				PosY: uint32(targetPy),
			})
			ai.aiMoveTarget = nil
			return
		}
	}

	// 定期发送移动更新（模拟客户端逐格上报）
	if now.Sub(ai.aiMoveLastUpdate) >= aiMoveUpdateInterval {
		current := ai.entity.GetPosition()
		if current != nil {
			currentPx, currentPy := argsdef.TileCoordToPixel(current.X, current.Y)
			speed := uint32(ai.entity.GetAttrSys().GetAttrValue(attrdef.AttrMoveSpeed))
			if speed <= 0 {
				speed = uint32(ai.moveStep())
			}
			moveSys.HandleUpdateMove(ai.scene, &protocol.C2SUpdateMoveReq{
				PosX:  uint32(currentPx),
				PosY:  uint32(currentPy),
				Speed: speed,
			})
			ai.aiMoveLastUpdate = now
		}
	}
}

func (ai *AISys) castSkill(target iface.IEntity) bool {
	fightSys, ok := ai.entity.GetFightSys().(*FightSys)
	if !ok {
		return false
	}
	skillID := ai.preferredSkill
	ctx := &argsdef.SkillCastContext{
		SkillId:   skillID,
		TargetHdl: target.GetHdl(),
		PosX:      target.GetPosition().X,
		PosY:      target.GetPosition().Y,
	}

	result, errCode := fightSys.CastSkill(ctx)
	if result == nil {
		result = &skill.CastResult{ErrCode: errCode}
	}
	SendSkillCastAck(ai.scene, ai.entity, skillID, errCode, DefaultLatencyToleranceMs)
	if errCode != int(protocol.SkillUseErr_SkillUseErrSuccess) {
		return false
	}
	fightSys.ApplySkillHits(ai.scene, skillID, result.Hits)
	return true
}

func (ai *AISys) moveStep() float64 {
	speed := ai.entity.GetAttrSys().GetAttrValue(attrdef.AttrMoveSpeed)
	if speed <= 0 {
		return minStepDistance
	}
	step := float64(speed) / 5.0
	if step < minStepDistance {
		step = minStepDistance
	}
	return step
}

func (ai *AISys) randomPatrolPoint() *argsdef.Position {
	radius := float64(ai.aiConfig.PatrolRadius)
	angle := ai.rand.Float64() * 2 * math.Pi
	length := ai.rand.Float64() * radius
	x := clampToUint32(float64(ai.homePos.X) + math.Cos(angle)*length)
	y := clampToUint32(float64(ai.homePos.Y) + math.Sin(angle)*length)
	return &argsdef.Position{X: x, Y: y}
}

func (ai *AISys) distanceTo(pos *argsdef.Position) float64 {
	return distanceBetween(ai.entity.GetPosition(), pos)
}

func clampToUint32(val float64) uint32 {
	if val < 0 {
		return 0
	}
	if val > float64(^uint32(0)) {
		return ^uint32(0)
	}
	return uint32(val)
}

// distanceBetween 计算两个位置之间的格子距离（欧几里得距离）
// 注意：Position 中的 X、Y 是格子坐标，不是像素坐标
func distanceBetween(a, b *argsdef.Position) float64 {
	if a == nil || b == nil {
		return math.MaxFloat64
	}
	// 计算格子坐标的欧几里得距离
	dx := float64(int64(a.X) - int64(b.X))
	dy := float64(int64(a.Y) - int64(b.Y))
	return math.Hypot(dx, dy)
}
