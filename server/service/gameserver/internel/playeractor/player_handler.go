package playeractor

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/actorprotocol"
	"postapocgame/server/service/gameserver/internel/dungeonserverlink"
	"postapocgame/server/service/gameserver/internel/entity"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/manager"
	"time"
)

// PlayerHandler 玩家消息处理器
type PlayerHandler struct {
	actor.BaseActorMsgHandler
}

// NewPlayerHandler 创建玩家消息处理器
func NewPlayerHandler() *PlayerHandler {
	return &PlayerHandler{}
}

func (h *PlayerHandler) HandleActorMessage(msg *actor.Message) error {
	session := gatewaylink.GetSession(msg.SessionId)
	if session == nil {
		return customerr.NewCustomErr("not found %s session", msg.SessionId)
	}
	msgId := msg.MsgId
	switch msgId {
	case protocol.C2S_Verify:
		return h.handleVerify(msg)
	case protocol.C2S_QueryRoles:
		return h.handleQueryRoles(msg)
	case protocol.C2S_CreateRole:
		return h.handleCreateRole(msg)
	case protocol.C2S_EnterGame:
		return h.handleEnterGame(msg)
	case protocol.C2S_Reconnect:
		return h.handleReconnect(msg)
	default:
		roleId := session.GetRoleId()
		if roleId > 0 {
			playerRole := manager.GetPlayerRole(roleId)
			if playerRole == nil {
				return customerr.NewCustomErr("not found %s session %d player role %d", msg.SessionId, roleId)
			}
			var protoIdH, protoIdL = msgId >> 8, msgId & 0xff
			getFunc := actorprotocol.GetFunc(protoIdH, protoIdL)
			if getFunc == nil {
				return customerr.NewCustomErr("not found %d %d handler", protoIdH, protoIdL)
			}
			return getFunc(playerRole, msg)
		}
	}
	return nil
}

func (h *PlayerHandler) handleVerify(msg *actor.Message) error {
	return nil
}
func (h *PlayerHandler) handleQueryRoles(msg *actor.Message) error {
	log.Infof("handleQueryRoles: SessionId=%s", msg.SessionId)

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
	return gatewaylink.SendToSession(msg.SessionId, protocol.S2C_RoleList, jsonData)
}

func (h *PlayerHandler) handleEnterGame(msg *actor.Message) error {
	log.Infof("handleSelectRole: SessionId=%s", msg.SessionId)

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
	err = h.enterGame(msg.SessionId, selectedRole)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (h *PlayerHandler) handleReconnect(msg *actor.Message) error {
	return nil
}

func (h *PlayerHandler) handleCreateRole(msg *actor.Message) error {
	log.Infof("handleCreateRole: SessionId=%s", msg.SessionId)

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
	if err := gatewaylink.SendToSession(msg.SessionId, protocol.S2C_CreateRoleResult, jsonData); err != nil {
		return customerr.Wrap(err)
	}

	return nil
}

// enterGame 进入游戏
func (h *PlayerHandler) enterGame(sessionId string, roleInfo *protocol.RoleInfo) error {
	log.Infof("enterGame: SessionId=%s, RoleId=%d", sessionId, roleInfo.RoleId)

	// 创建PlayerRole实例
	playerRole := entity.NewPlayerRole(sessionId, roleInfo)

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
