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
	"time"
)

func handleVerify(sessionId string, msg *network.ClientMessage) error {
	return nil
}

func handleQueryRoles(sessionId string, msg *network.ClientMessage) error {
	log.Infof("handleQueryRoles: SessionId=%s", sessionId)

	// 模拟返回固定的两个角色
	roleList := &protocol.RoleListResponse{
		Roles: []*protocol.RoleInfo{
			{
				RoleId: 10001,
				Job:    1,
				Sex:    1,
				Name:   "战士001",
				Level:  10,
			},
			{
				RoleId: 10002,
				Job:    2,
				Sex:    0,
				Name:   "法师002",
				Level:  15,
			},
		},
	}

	// 序列化为JSON
	jsonData, err := tool.JsonMarshal(roleList)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 发送给客户端
	return gatewaylink.SendToSession(sessionId, protocol.S2C_RoleList, jsonData)
}

func handleEnterGame(sessionId string, msg *network.ClientMessage) error {
	log.Infof("handleSelectRole: SessionId=%s", sessionId)

	// 解析选择角色请求
	req, err := protocol.UnmarshalSelectRoleRequest(msg.Data)
	if err != nil {
		log.Errorf("unmarshal select player role request failed: %v", err)
		return err
	}

	// 从模拟数据中查找角色
	var selectedRole *protocol.RoleInfo
	if req.RoleId == 10001 {
		selectedRole = &protocol.RoleInfo{
			RoleId: 10001,
			Job:    1,
			Sex:    1,
			Name:   "战士001",
			Level:  10,
		}
	} else if req.RoleId == 10002 {
		selectedRole = &protocol.RoleInfo{
			RoleId: 10002,
			Job:    2,
			Sex:    0,
			Name:   "法师002",
			Level:  15,
		}
	} else {
		return customerr.NewCustomErr("not found role Id")
	}

	log.Infof("Selected player role: RoleId=%d, Name=%s", selectedRole.RoleId, selectedRole.Name)

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
	log.Infof("handleCreateRole: SessionId=%s", sessionId)

	// 解析创建角色请求
	var req protocol.CreateRoleRequest
	err := tool.JsonUnmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("unmarshal create player role request failed: %v", err)
		return err
	}

	// 生成新角色
	roleId := time.Now().UnixNano()
	newRole := &protocol.RoleInfo{
		RoleId: uint64(roleId),
		Job:    req.Job,
		Sex:    req.Sex,
		Name:   req.Name,
		Level:  1,
	}

	// 创建响应
	resp := &protocol.CreateRoleResponse{
		Success: true,
		Role:    newRole,
	}

	// 序列化为JSON
	jsonData, err := tool.JsonMarshal(resp)
	if err != nil {
		log.Errorf("marshal create player role response failed: %v", err)
		return customerr.Wrap(err)
	}

	// 发送给客户端
	if err := gatewaylink.SendToSession(sessionId, protocol.S2C_CreateRoleResult, jsonData); err != nil {
		return customerr.Wrap(err)
	}

	return nil
}

// enterGame 进入游戏
func enterGame(sessionId string, roleInfo *protocol.RoleInfo) error {
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
	err = dungeonserverlink.AsyncCall(context.Background(), 1, sessionId, protocol.RPC_EnterDungeon, reqData)
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
	case protocol.C2S_Verify:
		err = handleVerify(sessionId, cliMsg)
	case protocol.C2S_QueryRoles:
		err = handleQueryRoles(sessionId, cliMsg)
	case protocol.C2S_CreateRole:
		err = handleCreateRole(sessionId, cliMsg)
	case protocol.C2S_EnterGame:
		err = handleEnterGame(sessionId, cliMsg)
	case protocol.C2S_Reconnect:
		err = handleReconnect(sessionId, cliMsg)
	default:
		roleId := session.GetRoleId()
		if roleId > 0 {
			playerRole := manager.GetPlayerRole(roleId)
			if playerRole != nil {
				var protoIdH, protoIdL = cliMsg.MsgId >> 8, cliMsg.MsgId & 0xff
				getFunc := clientprotocol.GetFunc(protoIdH, protoIdL)
				if getFunc != nil {
					err = getFunc(playerRole, cliMsg)
				}
			}
		}
	}
	if err != nil {
		log.Errorf("err:%v", err)
		return
	}
	return
}

func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		gshare.PlayerRegisterHandler(gshare.DoNetWorkMsg, handleDoNetWorkMsg)
	})
}
