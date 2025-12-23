package controller

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/playeractor/entity"
	"postapocgame/server/service/gameserver/internel/playeractor/router"
	"postapocgame/server/service/gameserver/internel/playeractor/skill"

	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"

	"google.golang.org/protobuf/proto"
)

// HandleEnterGame 处理进入游戏
func HandleEnterGame(ctx context.Context, msg *network.ClientMessage) error {
	var req protocol.C2SEnterGameReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	sessionId, err := sessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	depsSet := resolveEnterGameDeps(ctx)
	return depsSet.enterGame(ctx, sessionId, &req)
}

type enterGameDeps struct {
	roleRepo iface.RoleRepository
	roleMgr  iface.IPlayerRoleManager
}

func resolveEnterGameDeps(ctx context.Context) enterGameDeps {
	egd := enterGameDeps{
		roleRepo: deps.NewRoleRepository(),
		roleMgr:  deps.GetPlayerRoleManager(),
	}

	if rt := deps.FromContext(ctx); rt != nil {
		if rt.RoleRepo() != nil {
			egd.roleRepo = rt.RoleRepo()
		}
	}
	return egd
}

func (d enterGameDeps) enterGame(ctx context.Context, sessionId string, req *protocol.C2SEnterGameReq) error {
	session := gatewaylink.GetSession(sessionId)
	if session == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session not found")
	}

	role, err := d.roleRepo.GetRoleByID(ctx, req.RoleId)
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

	playerRole := entity.NewPlayerRole(sessionId, selectedRole)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "create player role failed")
	}

	d.roleMgr.Add(playerRole)
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
	sessionId, err := sessionIDFromContext(message.GetContext())
	if err != nil {
		return
	}
	session := gatewaylink.GetSession(sessionId)
	if session == nil {
		return
	}
	if iPlayerRole := deps.GetPlayerRoleManager().GetBySession(sessionId); iPlayerRole != nil {
		iPlayerRole.RunOne()
	}
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
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		protocolRouter := router.NewProtocolRouterController(nil)
		gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PAMNetworkMsg), protocolRouter.HandleDoNetworkMsg)
		gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PAMRunOneMsg), HandleRunOneMsg)
		gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PAMSendToClient), HandleSendToClient)
		router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SEnterGame), HandleEnterGame)
	})
}

func sessionIDFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "context missing")
	}
	sessionId, _ := ctx.Value(gshare.ContextKeySession).(string)
	if sessionId == "" {
		return "", customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session not found")
	}
	return sessionId, nil
}
