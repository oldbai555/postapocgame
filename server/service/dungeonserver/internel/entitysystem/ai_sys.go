package entitysystem

import (
	"math"
	"math/rand"
	"time"

	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/dungeonserver/internel/entitymgr"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"postapocgame/server/service/dungeonserver/internel/skill"
)

const (
	DefaultMonsterSkillID uint32 = 1001
	minStepDistance              = 6.0
)

// AISys 怪物AI系统
type AISys struct {
	entity         iface.IEntity
	scene          iface.IScene
	aiConfig       jsonconf.MonsterAIConfig
	state          AIState
	target         iface.IEntity
	homePos        argsdef.Position
	patrolTarget   *argsdef.Position
	rand           *rand.Rand
	lastAttack     time.Time
	preferredSkill uint32
	thinkInterval  time.Duration
	nextThink      time.Time
}

// NewAISys 创建AI系统
func NewAISys(entity iface.IEntity, monsterCfg *jsonconf.MonsterConfig) *AISys {
	if entity == nil || monsterCfg == nil {
		return nil
	}
	cfg := monsterCfg.AIConfig.WithDefaults()
	stepRand := rand.New(rand.NewSource(time.Now().UnixNano()))

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
	}
	if ai.patrolTarget != nil {
		if ai.moveTowards(ai.patrolTarget) {
			ai.patrolTarget = nil
		}
		ai.state = AIStatePatrol
	}
}

func (ai *AISys) handleChase() {
	if ai.target == nil {
		ai.state = AIStateReturning
		return
	}
	dist := ai.distanceTo(ai.target.GetPosition())
	if dist <= float64(ai.aiConfig.AttackRange) {
		ai.state = AIStateAttack
		return
	}
	ai.moveTowards(ai.target.GetPosition())
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
		ai.lastAttack = time.Now()
	}
}

func (ai *AISys) handleReturn() {
	dist := ai.distanceTo(&ai.homePos)
	if dist <= 5 {
		ai.state = AIStateIdle
		ai.target = nil
		return
	}
	ai.moveTowards(&ai.homePos)
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
	if target == nil {
		return false
	}
	moveSys := ai.entity.GetMoveSys()
	if moveSys == nil {
		return false
	}
	speed := float64(ai.entity.GetAttrSys().GetAttrValue(attrdef.AttrMoveSpeed))
	if speed <= 0 {
		speed = ai.moveStep()
	}
	moveSys.MoveTo(target, speed)
	return ai.distanceTo(target) < 3
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
	if errCode != int(protocol.SkillUseErr_SkillUseErrSuccess) || result == nil {
		return false
	}

	ai.broadcastSkillResult(skillID, result, errCode)
	return true
}

func (ai *AISys) broadcastSkillResult(skillID uint32, result *skill.CastResult, errCode int) {
	if ai.scene == nil {
		return
	}
	resp := &protocol.S2CSkillCastResultReq{
		CasterHdl: ai.entity.GetHdl(),
		SkillId:   skillID,
		ErrCode:   uint32(errCode),
		Hits:      convertAIHitResults(result.HitResults),
	}
	for _, et := range ai.scene.GetAllEntities() {
		_ = et.SendProtoMessage(uint16(protocol.S2CProtocol_S2CSkillCastResult), resp)
	}
}

func convertAIHitResults(hits []*skill.SkillHitResult) []*protocol.SkillHitResultSt {
	if len(hits) == 0 {
		return nil
	}
	protoHits := make([]*protocol.SkillHitResultSt, 0, len(hits))
	for _, hit := range hits {
		if hit == nil {
			continue
		}
		protoHit := &protocol.SkillHitResultSt{
			TargetHdl:  hit.TargetHdl,
			IsHit:      hit.IsHit,
			IsDodge:    hit.IsDodge,
			IsCrit:     hit.IsCrit,
			Damage:     hit.Damage,
			Heal:       hit.Heal,
			AddedBuffs: hit.AddedBuffs,
			ResultType: uint32(hit.ResultType),
			Attrs:      hit.Attrs,
			StateFlags: hit.StateFlags,
		}
		protoHits = append(protoHits, protoHit)
	}
	return protoHits
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

func distanceBetween(a, b *argsdef.Position) float64 {
	if a == nil || b == nil {
		return math.MaxFloat64
	}
	dx := float64(int64(a.X) - int64(b.X))
	dy := float64(int64(a.Y) - int64(b.Y))
	return math.Hypot(dx, dy)
}
