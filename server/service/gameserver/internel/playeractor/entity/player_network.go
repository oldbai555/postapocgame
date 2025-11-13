/**
 * @Author: zjj
 * @Date: 2025/11/11
 * @Desc:
**/

package entity

import (
	"context"
	"fmt"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/base"
	"postapocgame/server/service/gameserver/internel/clientprotocol"
	"postapocgame/server/service/gameserver/internel/dungeonserverlink"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/manager"
)

func handleVerify(ctx context.Context, msg *network.ClientMessage) error {
	return nil
}

func handleQueryRoles(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	log.Infof("handleQueryRoles: SessionId=%s", sessionId)

	// 模拟返回固定的两个角色
	roleList := &protocol.S2CRoleListReq{
		RoleList: []*protocol.PlayerSimpleData{
			{
				RoleId:   10001,
				Job:      1,
				Sex:      1,
				RoleName: "战士001",
				Level:    10,
			},
			{
				RoleId:   10002,
				Job:      2,
				Sex:      0,
				RoleName: "法师002",
				Level:    15,
			},
		},
	}

	// 序列化为JSON
	jsonData, err := tool.JsonMarshal(roleList)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 发送给客户端
	return gatewaylink.SendToSession(sessionId, uint16(protocol.S2CProtocol_S2CRoleList), jsonData)
}

func handleEnterGame(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	log.Infof("handleSelectRole: SessionId=%s", sessionId)

	// 解析选择角色请求
	var req protocol.C2SEnterGameReq
	err := tool.JsonUnmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("unmarshal select player role request failed: %v", err)
		return err
	}

	// 从模拟数据中查找角色
	var selectedRole *protocol.PlayerSimpleData
	if req.RoleId == 10001 {
		selectedRole = &protocol.PlayerSimpleData{
			RoleId:   10001,
			Job:      1,
			Sex:      1,
			RoleName: "战士001",
			Level:    10,
		}
	} else if req.RoleId == 10002 {
		selectedRole = &protocol.PlayerSimpleData{
			RoleId:   10002,
			Job:      2,
			Sex:      0,
			RoleName: "法师002",
			Level:    15,
		}
	} else {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "not found role Id")
	}

	log.Infof("Selected player role: RoleId=%d, Name=%s", selectedRole.RoleId, selectedRole.RoleName)

	// 进入游戏
	err = enterGame(sessionId, selectedRole)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func handleReconnect(ctx context.Context, msg *network.ClientMessage) error {
	return nil
}

func handleCreateRole(ctx context.Context, msg *network.ClientMessage) error {
	return nil
}

// enterGame 进入游戏
func enterGame(sessionId string, roleInfo *protocol.PlayerSimpleData) error {
	log.Infof("enterGame: SessionId=%s, RoleId=%d", sessionId, roleInfo.RoleId)

	// 创建PlayerRole实例
	playerRole := NewPlayerRole(sessionId, roleInfo)

	// 添加到PlayerRole管理器
	manager.GetPlayerRoleManager().Add(playerRole)
	session := gatewaylink.GetSession(sessionId)
	session.SetRoleId(playerRole.GetPlayerRoleId())

	// 构造进入副本请求
	reqData, err := tool.JsonMarshal(roleInfo)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 使用带SessionId的异步RPC调用
	err = dungeonserverlink.AsyncCall(context.Background(), uint8(protocol.SrvType_SrvTypeDungeonServer), sessionId, uint16(protocol.G2DRpcProtocol_G2DEnterDungeon), reqData)
	if err != nil {
		log.Errorf("call dungeon service enter scene failed: %v", err)
		return customerr.Wrap(err, int32(protocol.ErrorCode_Internal_Error))
	}

	err = playerRole.OnLogin()
	if err != nil {
		return customerr.Wrap(err)
	}

	return nil
}

func handleDoNetWorkMsg(message actor.IActorMessage) {
	msg, ok := message.(*base.SessionMessage)
	if !ok {
		return
	}

	sessionId := msg.SessionId

	session := gatewaylink.GetSession(sessionId)
	if session == nil {
		return
	}

	cliMsg, err := network.DefaultCodec().DecodeClientMessage(message.GetData())
	if err != nil {
		log.Errorf("err:%v", err)
		return
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, gshare.ContextKeySession, sessionId)
	var buildPlayerRoleCtx = func(ctx context.Context, roleId uint64) context.Context {
		if roleId == 0 {
			return ctx
		}
		pr := manager.GetPlayerRole(roleId)
		if pr != nil {
			return ctx
		}
		return pr.WithContext(ctx)
	}
	getFunc := clientprotocol.GetFunc(cliMsg.MsgId)
	if getFunc == nil {
		err = customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), fmt.Sprintf("not found %d handler", cliMsg.MsgId))
	}
	if err == nil {
		roleId := session.GetRoleId()
		ctx = buildPlayerRoleCtx(ctx, roleId)
		err = getFunc(ctx, cliMsg)
	}
	if err == nil {
		return
	}

	log.Errorf("handleDoNetWorkMsg failed, err:%v", err)
	err = gatewaylink.SendToSessionJSON(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
		Code: -1,
		Msg:  err.Error(),
	})
	if err != nil {
		log.Errorf("err:%v", err)
	}
}

func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		gshare.RegisterHandler(gshare.DoNetWorkMsg, handleDoNetWorkMsg)

		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SVerify), handleVerify)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SQueryRoles), handleQueryRoles)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SCreateRole), handleCreateRole)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SEnterGame), handleEnterGame)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SReconnect), handleReconnect)
	})
}
