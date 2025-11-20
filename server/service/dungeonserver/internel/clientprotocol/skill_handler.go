package clientprotocol

import (
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
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
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
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

	entitysystem.SendSkillCastAck(scene, entity, skillId, errCode, entitysystem.DefaultLatencyToleranceMs)

	if errCode == int(protocol.SkillUseErr_SkillUseErrSuccess) {
		fightSys.ApplySkillHits(scene, skillId, result.Hits)
	}

	return nil
}
