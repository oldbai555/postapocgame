package dungeonactor

import (
	"context"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/dungeonactor/entity"
	"postapocgame/server/service/gameserver/internel/dungeonactor/entitymgr"
	"postapocgame/server/service/gameserver/internel/dungeonactor/fbmgr"
	"postapocgame/server/service/gameserver/internel/gshare"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
)

// handleEnterGame 处理 PlayerActor → DungeonActor 的进入游戏请求
// 入口：protocol.DungeonActorMsgId_DAMEnterGame
func handleEnterGame(msg actor.IActorMessage) error {
	if msg == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "nil message")
	}

	ctx := msg.GetContext()
	if ctx == nil {
		ctx = context.Background()
	}

	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil || playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role missing")
	}
	roleData := playerRole.GetPlayerSimpleData()
	if roleData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "role data missing")
	}
	sessionID := playerRole.GetSessionId()
	if sessionID == "" {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session id missing")
	}

	var req protocol.DAMEnterGameReq
	if raw := msg.GetData(); len(raw) > 0 {
		if err := proto.Unmarshal(raw, &req); err != nil {
			return customerr.Wrap(err)
		}
	}

	fb, ok := fbmgr.GetFuBenMgr().GetFuBen(0)
	if !ok || fb == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "default fuben missing")
	}
	scenes := fb.GetAllScenes()
	if len(scenes) == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "default scenes missing")
	}
	scene := scenes[0]

	spawnX, spawnY := scene.GetSpawnPos()

	player := entity.NewPlayer(sessionID, roleData, req.SkillMap)
	player.SetPosition(spawnX, spawnY)

	if err := fb.OnPlayerEnter(sessionID); err != nil {
		return customerr.Wrap(err)
	}
	if err := scene.AddEntity(player); err != nil {
		return customerr.Wrap(err)
	}
	entitymgr.GetEntityMgr().BindSession(sessionID, player.GetHdl())

	enterScene := &protocol.S2CEnterSceneReq{
		EntityData: player.BuildProtoEntitySt(),
	}
	if err := player.SendProtoMessage(uint16(protocol.S2CProtocol_S2CEnterScene), enterScene); err != nil {
		log.Warnf("[dungeon-actor] send enter scene failed: %v", err)
	}

	for _, et := range scene.GetAllEntities() {
		if et == nil || et.GetHdl() == player.GetHdl() {
			continue
		}

		appear := &protocol.S2CEntityAppearReq{
			Entity: et.BuildProtoEntitySt(),
		}
		if err := player.SendProtoMessage(uint16(protocol.S2CProtocol_S2CEntityAppear), appear); err != nil {
			log.Warnf("[dungeon-actor] notify appear to player failed: %v", err)
		}

		if et.GetEntityType() == uint32(protocol.EntityType_EtPlayer) {
			back := &protocol.S2CEntityAppearReq{Entity: player.BuildProtoEntitySt()}
			_ = et.SendProtoMessage(uint16(protocol.S2CProtocol_S2CEntityAppear), back)
		}
	}

	return nil
}
