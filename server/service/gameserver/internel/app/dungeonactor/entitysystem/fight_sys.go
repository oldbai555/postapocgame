/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package entitysystem

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entitymgr"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/iface"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/skill"
)

var _ iface.IFightSys = (*FightSys)(nil)

// FightSys 管理实体主动/被动技能，并在命中后驱动伤害、治疗与 Buff 结算。
type FightSys struct {
	et              iface.IEntity
	skills          map[uint32]*skill.Skill //手动释放技能列表
	passivitySkills map[uint32]*skill.Skill //被动技能列表

	CommonCd map[uint32]int64 // 技能组公共CD
}

func NewFightSys() *FightSys {
	return &FightSys{
		skills:          make(map[uint32]*skill.Skill),
		passivitySkills: make(map[uint32]*skill.Skill),
		CommonCd:        make(map[uint32]int64),
	}
}

func (s *FightSys) SetEntity(et iface.IEntity) {
	s.et = et
}

func (s *FightSys) LearnSkill(skillId, skillLv uint32) error {
	configMgr := jsonconf.GetConfigManager()
	skillCfg := configMgr.GetSkillConfig(skillId)
	if skillCfg == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "skill config not found:%d", skillId)
	}
	sk := skill.NewSkill(skillId, skillLv)
	sk.SetCd(0)
	s.skills[skillId] = sk
	return nil
}

func (s *FightSys) HasSkill(skillId uint32) bool {
	_, ok := s.skills[skillId]
	return ok
}

func (s *FightSys) UseSkill(ctx *argsdef.SkillCastContext) int {
	_, errCode := s.CastSkill(ctx)
	return errCode
}

func (s *FightSys) CastSkill(ctx *argsdef.SkillCastContext) (*skill.CastResult, int) {
	caster := s.et
	log.Infof("=== Skill Cast Start === Caster=%d, SkillId=%d", caster.GetHdl(), ctx.SkillId)

	skillId := ctx.SkillId
	sk := s.skills[skillId]
	if sk == nil {
		return &skill.CastResult{ErrCode: int(protocol.SkillUseErr_ErrSkillNotLearned)}, int(protocol.SkillUseErr_ErrSkillNotLearned)
	}

	if !sk.CheckCd() {
		return &skill.CastResult{ErrCode: int(protocol.SkillUseErr_ErrSkillInCooldown)}, int(protocol.SkillUseErr_ErrSkillInCooldown)
	}

	skillCfg := sk.GetConfig()
	if skillCfg == nil {
		return &skill.CastResult{ErrCode: int(protocol.SkillUseErr_ErrSkillNotLearned)}, int(protocol.SkillUseErr_ErrSkillNotLearned)
	}

	if caster.GetMP() < int64(skillCfg.ManaCost) {
		return &skill.CastResult{ErrCode: int(protocol.SkillUseErr_ErrSkillNotEnoughMP)}, int(protocol.SkillUseErr_ErrSkillNotEnoughMP)
	}

	result := sk.Use(ctx, caster)
	if result == nil {
		return &skill.CastResult{ErrCode: int(protocol.SkillUseErr_ErrSkillCannotCast)}, int(protocol.SkillUseErr_ErrSkillCannotCast)
	}
	return result, result.ErrCode
}

func (s *FightSys) RunOne(time.Time) {}

func (s *FightSys) ApplySkillHits(scene iface.IScene, skillId uint32, hits []*skill.SkillHitResult) {
	if s == nil || scene == nil || len(hits) == 0 {
		return
	}
	s.applyHitEffects(hits)
	resp := &protocol.S2CSkillDamageResultReq{
		CasterHdl:    s.et.GetHdl(),
		SkillId:      skillId,
		BatchIndex:   0,
		ServerTimeMs: servertime.UnixMilli(),
		Hits:         ConvertSkillHitsToProto(hits),
	}
	BroadcastSceneProto(scene, uint16(protocol.S2CProtocol_S2CSkillDamageResult), resp)
}

func (s *FightSys) applyHitEffects(hits []*skill.SkillHitResult) {
	caster := s.et
	for _, hit := range hits {
		if hit == nil || hit.Target == nil {
			continue
		}
		target := hit.Target
		switch hit.ResultType {
		case skill.SkillResultTypeDamage:
			if hit.Damage > 0 {
				target.OnAttacked(caster, hit.Damage)
			}
		case skill.SkillResultTypeHeal:
			if hit.Heal > 0 {
				current := target.GetHP() + hit.Heal
				if current > target.GetMaxHP() {
					current = target.GetMaxHP()
				}
				target.SetHP(current)
			}
		}
		if len(hit.AddedBuffs) > 0 {
			if buffSys := target.GetBuffSys(); buffSys != nil {
				for _, buffId := range hit.AddedBuffs {
					if err := buffSys.AddBuff(buffId, caster); err != nil {
						log.Errorf("AddBuff failed err:%v", err)
					}
				}
			}
		}
	}
}

const normalAttackSkillID = 1001

// handleUseSkill 处理技能使用请求（Actor 消息版本）
// 约定：msg.Context 中包含 "session" 字段，可通过 EntityMgr 根据 sessionId 查到实体。
func HandleUseSkill(msg actor.IActorMessage) error {
	if msg == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "nil message")
	}

	ctx := msg.GetContext()
	if ctx == nil {
		ctx = context.Background()
	}
	sessionId, _ := ctx.Value("session").(string)
	if sessionId == "" {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "not found session")
	}

	// 获取实体
	entityAny, ok := entitymgr.GetEntityMgr().GetBySession(sessionId)
	if !ok || entityAny == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "entity not found for session")
	}
	entity := entityAny
	if entity == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "invalid entity type")
	}

	var req protocol.C2SUseSkillReq
	if err := proto.Unmarshal(msg.GetData(), &req); err != nil {
		return err
	}

	scene, err := getSceneByEntity(entity)
	if err != nil {
		return err
	}

	skillId := req.SkillId
	if skillId == 0 {
		skillId = normalAttackSkillID
	}

	// 将客户端发送的像素坐标转换为格子坐标
	tileX, tileY := argsdef.PixelCoordToTile(req.PosX, req.PosY)
	skillCtx := &argsdef.SkillCastContext{
		TargetHdl:     req.TargetHdl,
		TargetHdlList: req.TargetList,
		SkillId:       skillId,
		PosX:          tileX, // 格子坐标
		PosY:          tileY, // 格子坐标
	}

	fightSys, ok := entity.GetFightSys().(*FightSys)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "invalid fight sys type")
	}

	result, errCode := fightSys.CastSkill(skillCtx)
	if result == nil {
		result = &skill.CastResult{
			ErrCode: errCode,
		}
	}

	SendSkillCastAck(scene, entity, skillId, errCode, DefaultLatencyToleranceMs)

	if errCode == int(protocol.SkillUseErr_SkillUseErrSuccess) {
		fightSys.ApplySkillHits(scene, skillId, result.Hits)
	}

	return nil
}

// getSceneByEntity 获取实体所在场景
func getSceneByEntity(entity iface.IEntity) (iface.IScene, error) {
	if entity == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "entity missing")
	}
	scene, ok := entitymgr.GetEntityMgr().GetSceneByHandle(entity.GetHdl())
	if !ok {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "scene not bound")
	}
	return scene, nil
}
