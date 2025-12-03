// skill_util.go 汇总技能广播与命中结果转换的工具函数，供战斗系统复用。
package entitysystem

import (
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/iface"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/skill"
)

const DefaultLatencyToleranceMs = 150

// SendSkillCastAck 根据 errCode 返回给施法者或全场
func SendSkillCastAck(scene iface.IScene, caster iface.IEntity, skillId uint32, errCode int, toleranceMs uint32) {
	if toleranceMs == 0 {
		toleranceMs = DefaultLatencyToleranceMs
	}
	resp := &protocol.S2CSkillCastResultReq{
		CasterHdl:          caster.GetHdl(),
		SkillId:            skillId,
		ErrCode:            uint32(errCode),
		ServerTimeMs:       servertime.UnixMilli(),
		LatencyToleranceMs: toleranceMs,
	}

	if errCode == int(protocol.SkillUseErr_SkillUseErrSuccess) {
		BroadcastSceneProto(scene, uint16(protocol.S2CProtocol_S2CSkillCastResult), resp)
		return
	}
	if caster != nil {
		_ = caster.SendProtoMessage(uint16(protocol.S2CProtocol_S2CSkillCastResult), resp)
	}
}

// ConvertSkillHitsToProto 将命中结果转换为协议结构
func ConvertSkillHitsToProto(hits []*skill.SkillHitResult) []*protocol.SkillHitResultSt {
	if len(hits) == 0 {
		return nil
	}
	out := make([]*protocol.SkillHitResultSt, 0, len(hits))
	for _, hit := range hits {
		if hit == nil {
			continue
		}
		targetAttrs := buildTargetAttrs(hit.Target)
		stateFlags := buildTargetState(hit.Target)

		out = append(out, &protocol.SkillHitResultSt{
			TargetHdl:   hit.TargetHdl,
			IsHit:       hit.IsHit,
			IsDodge:     hit.IsDodge,
			IsCrit:      hit.IsCrit,
			Damage:      hit.Damage,
			Heal:        hit.Heal,
			AddedBuffs:  hit.AddedBuffs,
			ResultType:  protocol.SkillResultType(hit.ResultType),
			Attrs:       targetAttrs,
			StateFlags:  stateFlags,
			DamageFlags: hit.DamageFlags,
		})
	}
	return out
}

func buildTargetAttrs(target iface.IEntity) map[uint32]int64 {
	if target == nil {
		return nil
	}
	attrSys := target.GetAttrSys()
	if attrSys == nil {
		return nil
	}
	attrs := make(map[uint32]int64)
	copyAttrs := func(src map[attrdef.AttrType]attrdef.AttrValue) {
		for k, v := range src {
			attrs[uint32(k)] = int64(v)
		}
	}
	copyAttrs(attrSys.GetAllCombatAttrs())
	copyAttrs(attrSys.GetAllExtraAttrs())
	if len(attrs) == 0 {
		return nil
	}
	return attrs
}

func buildTargetState(target iface.IEntity) uint64 {
	if target == nil {
		return 0
	}
	return target.GetStateFlags()
}
