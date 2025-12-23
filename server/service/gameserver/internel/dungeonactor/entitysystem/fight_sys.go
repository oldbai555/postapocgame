/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package entitysystem

import (
	"context"
	"postapocgame/server/service/gameserver/internel/dungeonactor/entitymgr"
	iface2 "postapocgame/server/service/gameserver/internel/dungeonactor/iface"
	skill2 "postapocgame/server/service/gameserver/internel/dungeonactor/skill"
	"time"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
)

var _ iface2.IFightSys = (*FightSys)(nil)

// FightSys 管理实体主动/被动技能，并在命中后驱动伤害、治疗与 Buff 结算。
type FightSys struct {
	et              iface2.IEntity
	skills          map[uint32]*skill2.Skill //手动释放技能列表
	passivitySkills map[uint32]*skill2.Skill //被动技能列表

	CommonCd map[uint32]int64 // 技能组公共CD
}

func NewFightSys() *FightSys {
	return &FightSys{
		skills:          make(map[uint32]*skill2.Skill),
		passivitySkills: make(map[uint32]*skill2.Skill),
		CommonCd:        make(map[uint32]int64),
	}
}

func (s *FightSys) SetEntity(et iface2.IEntity) {
	s.et = et
}

func (s *FightSys) LearnSkill(skillId, skillLv uint32) error {
	configMgr := jsonconf.GetConfigManager()
	skillCfg := configMgr.GetSkillConfig(skillId)
	if skillCfg == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "skill config not found:%d", skillId)
	}
	sk := skill2.NewSkill(skillId, skillLv)
	sk.SetCd(0)
	s.skills[skillId] = sk
	return nil
}

func (s *FightSys) HasSkill(skillId uint32) bool {
	_, ok := s.skills[skillId]
	return ok
}

func (s *FightSys) UseSkill(ctx *argsdef.SkillCastContext) int {
	caster := s.et
	log.Infof("=== Skill Cast Start === Caster=%d, SkillId=%d", caster.GetHdl(), ctx.SkillId)

	skillId := ctx.SkillId
	sk := s.skills[skillId]
	if sk == nil {
		return int(protocol.SkillUseErr_ErrSkillNotLearned)
	}

	if !sk.CheckCd() {
		return int(protocol.SkillUseErr_ErrSkillInCooldown)
	}

	skillCfg := sk.GetConfig()
	if skillCfg == nil {
		return int(protocol.SkillUseErr_ErrSkillNotLearned)
	}

	if caster.GetMP() < int64(skillCfg.ManaCost) {
		return int(protocol.SkillUseErr_ErrSkillNotEnoughMP)
	}

	result := sk.Use(ctx, caster)
	if result == nil {
		return int(protocol.SkillUseErr_ErrSkillCannotCast)
	}
	return result.ErrCode
}

func (s *FightSys) RunOne(time.Time) {}

const normalAttackSkillID = 1001

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

	skillId := req.SkillId
	if skillId == 0 {
		skillId = normalAttackSkillID
	}

	// 将客户端发送的像素坐标转换为格子坐标
	tileX, tileY := argsdef.PixelCoordToTile(req.PosX, req.PosY)
	skillCtx := &argsdef.SkillCastContext{
		TargetHdl: req.TargetHdl,
		SkillId:   skillId,
		PosX:      tileX, // 格子坐标
		PosY:      tileY, // 格子坐标
	}

	fightSys, ok := entity.GetFightSys().(*FightSys)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "invalid fight sys type")
	}

	fightSys.UseSkill(skillCtx)

	return nil
}

// getSceneByEntity 获取实体所在场景
func getSceneByEntity(entity iface2.IEntity) (iface2.IScene, error) {
	if entity == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "entity missing")
	}
	scene, ok := entitymgr.GetEntityMgr().GetSceneByHandle(entity.GetHdl())
	if !ok {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "scene not bound")
	}
	return scene, nil
}
