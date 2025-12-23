package controller

import (
	"context"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/playeractor/entity"
	"postapocgame/server/service/gameserver/internel/playeractor/router"
	"postapocgame/server/service/gameserver/internel/playeractor/skill"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

// HandleEnterGame 处理进入游戏
func HandleEnterGame(ctx context.Context, msg *network.ClientMessage) error {
	var req protocol.C2SEnterGameReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	session := gatewaylink.GetSession(sessionId)
	if session == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session not found")
	}

	// 从 context 获取 Runtime，然后获取 RoleRepository
	rt := deps.FromContext(ctx)

	var roleRepo iface.RoleRepository
	if rt != nil {
		// 优先通过 Runtime 获取 RoleRepository（符合依赖注入原则）
		roleRepo = rt.RoleRepo()
	}
	if roleRepo == nil {
		// 如果 Runtime 中没有 RoleRepository，使用 deps.NewRoleRepository 作为回退（bootstrapping 场景）
		roleRepo = deps.NewRoleRepository()
	}

	role, err := roleRepo.GetRoleByID(ctx, req.RoleId)
	if err != nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "角色不存在")
	}

	if role.AccountID != uint64(session.GetAccountID()) {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "角色不属于当前账号")
	}

	selectedRole := &protocol.PlayerSimpleData{
		RoleId:   role.ID,
		Job:      role.Job,
		Sex:      role.Sex,
		RoleName: role.RoleName,
		Level:    role.Level,
	}

	if err := enterGame(sessionId, selectedRole); err != nil {
		return err
	}
	return nil
}

func enterGame(sessionId string, roleInfo *protocol.PlayerSimpleData) error {
	playerRole := entity.NewPlayerRole(sessionId, roleInfo)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "create player role failed")
	}

	deps.GetPlayerRoleManager().Add(playerRole)
	session := gatewaylink.GetSession(sessionId)
	if session == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session not found")
	}
	session.SetRoleId(playerRole.GetPlayerRoleId())

	if err := playerRole.OnLogin(); err != nil {
		return customerr.Wrap(err)
	}

	roleCtx := playerRole.WithContext(context.Background())
	skillSys := skill.GetSkillSys(roleCtx)
	var skillMap = make(map[uint32]uint32)
	if skillSys != nil {
		if m, err := skillSys.GetSkillMap(roleCtx); err == nil {
			skillMap = m
		}
	}

	reqData, err := proto.Marshal(&protocol.DAMEnterGameReq{
		SessionId:  sessionId,
		PlatformId: gshare.GetPlatformId(),
		SrvId:      gshare.GetSrvId(),
		SkillMap:   skillMap,
	})
	if err != nil {
		return customerr.Wrap(err)
	}

	if err := playerRole.CallDungeonActor(roleCtx, uint16(protocol.DungeonActorMsgId_DAMEnterGame), reqData); err != nil {
		return customerr.Wrap(err, int32(protocol.ErrorCode_Internal_Error))
	}

	return nil
}

// HandleRunOneMsg 驱动玩家 Actor RunOne
func HandleRunOneMsg(message actor.IActorMessage) {
	sessionId := message.GetContext().Value(gshare.ContextKeySession).(string)
	session := gatewaylink.GetSession(sessionId)
	if session == nil {
		return
	}
	iPlayerRole := deps.GetPlayerRoleManager().GetBySession(sessionId)
	if iPlayerRole == nil {
		return
	}
	iPlayerRole.RunOne()
}

// HandleSendToClient 统一的 S2C 透传
func HandleSendToClient(message actor.IActorMessage) {
	var req protocol.PAMSendToClientReq
	if err := proto.Unmarshal(message.GetData(), &req); err != nil {
		log.Errorf("[player-network] handleSendToClient: unmarshal failed: %v", err)
		return
	}

	sessionID, _ := message.GetContext().Value(gshare.ContextKeySession).(string)
	if sessionID == "" {
		log.Warnf("[player-network] handleSendToClient: missing session id")
		return
	}

	if err := gatewaylink.SendToSession(sessionID, uint16(req.GetMsgId()), req.GetData()); err != nil {
		log.Errorf("[player-network] handleSendToClient: send failed: %v", err)
	}
}

func init() {
	protocolRouter := router.NewProtocolRouterController(nil)
	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PAMNetworkMsg), protocolRouter.HandleDoNetworkMsg)
	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PAMRunOneMsg), HandleRunOneMsg)
	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PAMSendToClient), HandleSendToClient)

	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SEnterGame), HandleEnterGame)
}
