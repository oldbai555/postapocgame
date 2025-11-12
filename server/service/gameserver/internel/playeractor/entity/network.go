/**
 * @Author: zjj
 * @Date: 2025/11/11
 * @Desc:
**/

package entity

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/base"
	"postapocgame/server/service/gameserver/internel/dungeonserverlink"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/manager"
	"postapocgame/server/service/gameserver/internel/playeractor/clientprotocol"
)

func handleVerify(sessionId string, msg *network.ClientMessage) error {
	return nil
}

func handleQueryRoles(sessionId string, msg *network.ClientMessage) error {
	log.Infof("handleQueryRoles: SessionId=%s", sessionId)

	// 模拟返回固定的两个角色
	roleList := &protocol.S2CRoleListReq{
		RoleList: []*protocol.PlayerRoleData{
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

func handleEnterGame(sessionId string, msg *network.ClientMessage) error {
	log.Infof("handleSelectRole: SessionId=%s", sessionId)

	// 解析选择角色请求
	var req protocol.C2SEnterGameReq
	err := tool.JsonUnmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("unmarshal select player role request failed: %v", err)
		return err
	}

	// 从模拟数据中查找角色
	var selectedRole *protocol.PlayerRoleData
	if req.RoleId == 10001 {
		selectedRole = &protocol.PlayerRoleData{
			RoleId:   10001,
			Job:      1,
			Sex:      1,
			RoleName: "战士001",
			Level:    10,
		}
	} else if req.RoleId == 10002 {
		selectedRole = &protocol.PlayerRoleData{
			RoleId:   10002,
			Job:      2,
			Sex:      0,
			RoleName: "法师002",
			Level:    15,
		}
	} else {
		return customerr.NewCustomErr("not found role Id")
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

func handleReconnect(sessionId string, msg *network.ClientMessage) error {
	return nil
}

func handleCreateRole(sessionId string, msg *network.ClientMessage) error {
	return nil
}

// enterGame 进入游戏
func enterGame(sessionId string, roleInfo *protocol.PlayerRoleData) error {
	log.Infof("enterGame: SessionId=%s, RoleId=%d", sessionId, roleInfo.RoleId)

	// 创建PlayerRole实例
	playerRole := NewPlayerRole(sessionId, roleInfo)

	// 添加到PlayerRole管理器
	manager.GetPlayerRoleManager().Add(playerRole)
	session := gatewaylink.GetSession(sessionId)
	session.SetRoleId(playerRole.GetPlayerRoleId())

	// 调用OnLogin，触发所有系统数据下发
	if err := playerRole.OnLogin(); err != nil {
		log.Errorf("PlayerRole.OnLogin failed, err:%v", err)
	}

	// 构造进入副本请求
	reqData, err := tool.JsonMarshal(roleInfo)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 使用带SessionId的异步RPC调用
	err = dungeonserverlink.AsyncCall(context.Background(), 1, sessionId, uint16(protocol.G2DRpcProtocol_G2DEnterDungeon), reqData)
	if err != nil {
		log.Errorf("call dungeon service enter scene failed: %v", err)
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

	switch cliMsg.MsgId {
	case uint16(protocol.C2SProtocol_C2SVerify):
		err = handleVerify(sessionId, cliMsg)
	case uint16(protocol.C2SProtocol_C2SQueryRoles):
		err = handleQueryRoles(sessionId, cliMsg)
	case uint16(protocol.C2SProtocol_C2SCreateRole):
		err = handleCreateRole(sessionId, cliMsg)
	case uint16(protocol.C2SProtocol_C2SEnterGame):
		err = handleEnterGame(sessionId, cliMsg)
	case uint16(protocol.C2SProtocol_C2SReconnect):
		err = handleReconnect(sessionId, cliMsg)
	default:
		var doClientProtocol = func(roleId uint64) error {
			if roleId == 0 {
				return customerr.NewCustomErr("roleId is zero")
			}
			playerRole := manager.GetPlayerRole(roleId)
			if playerRole != nil {
				return customerr.NewCustomErr("not found %d player role", roleId)
			}
			var protoIdH, protoIdL = cliMsg.MsgId >> 8, cliMsg.MsgId & 0xff
			getFunc := clientprotocol.GetFunc(protoIdH, protoIdL)
			if getFunc == nil {
				return customerr.NewCustomErr("not found %d %d handler", protoIdH, protoIdL)
			}
			return getFunc(playerRole, cliMsg)
		}
		roleId := session.GetRoleId()
		err = doClientProtocol(roleId)
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
	})
}
