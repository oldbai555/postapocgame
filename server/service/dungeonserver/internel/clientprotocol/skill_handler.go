package clientprotocol

import (
	"postapocgame/server/internal"
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/entitysystem"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"postapocgame/server/service/dungeonserver/internel/skill"
)

const normalAttackSkillID = 1001

func init() {
	Register(uint16(protocol.C2SProtocol_C2SUseSkill), handleUseSkill)
}

func handleUseSkill(entity iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SUseSkillReq
	if err := internal.Unmarshal(msg.Data, &req); err != nil {
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

	ctx := &argsdef.SkillCastContext{
		TargetHdl:     req.TargetHdl,
		TargetHdlList: req.TargetList,
		SkillId:       skillId,
		PosX:          req.PosX,
		PosY:          req.PosY,
	}

	fightSys, ok := entity.GetFightSys().(*entitysystem.FightSys)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "invalid fight sys type")
	}

	result, errCode := fightSys.CastSkill(ctx)
	if result == nil {
		result = &skill.CastResult{
			ErrCode: errCode,
		}
	}

	resp := &protocol.S2CSkillCastResultReq{
		CasterHdl: entity.GetHdl(),
		SkillId:   skillId,
		ErrCode:   uint32(errCode),
		Hits:      convertHitResults(result.HitResults),
	}

	if errCode == int(protocol.SkillUseErr_SkillUseErrSuccess) {
		broadcastSceneMessage(scene, uint16(protocol.S2CProtocol_S2CSkillCastResult), resp)
	} else {
		log.Warnf("skill cast failed: hdl=%d skill=%d err=%d", entity.GetHdl(), skillId, errCode)
		_ = entity.SendJsonMessage(uint16(protocol.S2CProtocol_S2CSkillCastResult), resp)
	}

	return nil
}

func convertHitResults(hits []*skill.SkillHitResult) []*protocol.SkillHitResultSt {
	if len(hits) == 0 {
		return nil
	}
	protoHits := make([]*protocol.SkillHitResultSt, 0, len(hits))
	for _, hit := range hits {
		if hit == nil {
			continue
		}
		protoHits = append(protoHits, &protocol.SkillHitResultSt{
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
		})
	}
	return protoHits
}
