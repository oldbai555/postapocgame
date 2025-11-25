package entitysystem

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/playeractor/clientprotocol"
)

// FriendSys 好友系统
type FriendSys struct {
	*BaseSystem
	data *protocol.SiFriendData
}

// NewFriendSys 创建好友系统
func NewFriendSys() iface.ISystem {
	return &FriendSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysFriend)),
	}
}

func (s *FriendSys) OnInit(ctx context.Context) {
	role, err := GetIPlayerRoleByContext(ctx)
	if err != nil || role == nil {
		return
	}
	bd := role.GetBinaryData()
	if bd.FriendData == nil {
		bd.FriendData = &protocol.SiFriendData{
			FriendList:        make([]uint64, 0),
			FriendRequestList: make([]uint64, 0),
		}
	}
	s.data = bd.FriendData
}

// GetFriendList 获取好友列表
func (s *FriendSys) GetFriendList() []uint64 {
	if s.data == nil {
		return nil
	}
	return s.data.FriendList
}

// GetFriendRequestList 获取好友申请列表
func (s *FriendSys) GetFriendRequestList() []uint64 {
	if s.data == nil {
		return nil
	}
	return s.data.FriendRequestList
}

// AddFriend 添加好友
func (s *FriendSys) AddFriend(friendId uint64) bool {
	if s.data == nil {
		return false
	}
	// 检查是否已经是好友
	for _, id := range s.data.FriendList {
		if id == friendId {
			return false
		}
	}
	s.data.FriendList = append(s.data.FriendList, friendId)
	return true
}

// RemoveFriend 移除好友
func (s *FriendSys) RemoveFriend(friendId uint64) bool {
	if s.data == nil {
		return false
	}
	for i, id := range s.data.FriendList {
		if id == friendId {
			s.data.FriendList = append(s.data.FriendList[:i], s.data.FriendList[i+1:]...)
			return true
		}
	}
	return false
}

// AddFriendRequest 添加好友申请
func (s *FriendSys) AddFriendRequest(requesterId uint64) bool {
	if s.data == nil {
		return false
	}
	// 检查是否已经申请过
	for _, id := range s.data.FriendRequestList {
		if id == requesterId {
			return false
		}
	}
	s.data.FriendRequestList = append(s.data.FriendRequestList, requesterId)
	return true
}

// RemoveFriendRequest 移除好友申请
func (s *FriendSys) RemoveFriendRequest(requesterId uint64) bool {
	if s.data == nil {
		return false
	}
	for i, id := range s.data.FriendRequestList {
		if id == requesterId {
			s.data.FriendRequestList = append(s.data.FriendRequestList[:i], s.data.FriendRequestList[i+1:]...)
			return true
		}
	}
	return false
}

// GetFriendSys 获取好友系统
func GetFriendSys(ctx context.Context) *FriendSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysFriend))
	if system == nil {
		return nil
	}
	sys := system.(*FriendSys)
	if sys == nil || !sys.IsOpened() {
		return nil
	}
	return sys
}

// handleAddFriend 处理添加好友请求
func handleAddFriend(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleAddFriend: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SAddFriendReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleAddFriend: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()
	roleInfo := playerRole.GetRoleInfo()
	if roleInfo == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "角色信息不存在",
		})
	}

	// 验证目标角色
	if req.TargetId == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "目标角色ID无效",
		})
	}

	if req.TargetId == roleId {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "不能添加自己为好友",
		})
	}

	// 检查是否已经是好友
	friendSys := GetFriendSys(ctx)
	if friendSys != nil {
		friendList := friendSys.GetFriendList()
		for _, friendId := range friendList {
			if friendId == req.TargetId {
				return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CAddFriendResult), &protocol.S2CAddFriendResultReq{
					Success:  false,
					Message:  "已经是好友",
					TargetId: req.TargetId,
				})
			}
		}
	}

	// 发送到 PublicActor 处理
	addFriendMsg := &protocol.AddFriendReqMsg{
		RequesterId:   roleId,
		TargetId:      req.TargetId,
		RequesterName: roleInfo.RoleName,
	}
	msgData, err := proto.Marshal(addFriendMsg)
	if err != nil {
		log.Errorf("handleAddFriend: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAddFriendReq), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleAddFriend: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// handleRespondFriendReq 处理响应好友申请
func handleRespondFriendReq(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleRespondFriendReq: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SRespondFriendReqReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleRespondFriendReq: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()
	friendSys := GetFriendSys(ctx)
	if friendSys == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "好友系统未初始化",
		})
	}

	// 检查是否有该申请
	requestList := friendSys.GetFriendRequestList()
	hasRequest := false
	for _, requesterId := range requestList {
		if requesterId == req.RequesterId {
			hasRequest = true
			break
		}
	}

	if !hasRequest {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRespondFriendReqResult), &protocol.S2CRespondFriendReqResultReq{
			Success:     false,
			Message:     "未找到该好友申请",
			RequesterId: req.RequesterId,
			Accepted:    req.Accepted,
		})
	}

	// 移除申请
	friendSys.RemoveFriendRequest(req.RequesterId)

	// 如果同意，添加好友
	if req.Accepted {
		friendSys.AddFriend(req.RequesterId)
	}

	// 如果同意，需要通知申请者，并让申请者也添加目标为好友
	if req.Accepted {
		// 发送响应消息到 PublicActor，通知申请者
		respMsg := &protocol.AddFriendRespMsg{
			RequesterId: req.RequesterId,
			TargetId:    roleId,
			Accepted:    true,
		}
		msgData, err := proto.Marshal(respMsg)
		if err != nil {
			log.Errorf("handleRespondFriendReq: marshal failed: %v", err)
			return customerr.Wrap(err)
		}

		actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAddFriendResp), msgData)
		err = gshare.SendPublicMessageAsync("global", actorMsg)
		if err != nil {
			log.Errorf("handleRespondFriendReq: send to public actor failed: %v", err)
			return customerr.Wrap(err)
		}
	}

	// 返回结果给客户端
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRespondFriendReqResult), &protocol.S2CRespondFriendReqResultReq{
		Success:     true,
		Message:     "操作成功",
		RequesterId: req.RequesterId,
		Accepted:    req.Accepted,
	})
}

// handleQueryFriendList 处理查询好友列表
func handleQueryFriendList(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleQueryFriendList: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	roleId := playerRole.GetPlayerRoleId()
	_ = roleId // 暂时未使用，后续完善时使用
	friendSys := GetFriendSys(ctx)
	if friendSys == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CFriendList), &protocol.S2CFriendListReq{
			Friends:      []*protocol.PlayerRankSnapshot{},
			OnlineStatus: make(map[uint64]bool),
		})
	}

	// 获取好友列表
	friendList := friendSys.GetFriendList()
	if len(friendList) == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CFriendList), &protocol.S2CFriendListReq{
			Friends:      []*protocol.PlayerRankSnapshot{},
			OnlineStatus: make(map[uint64]bool),
		})
	}

	// 发送到 PublicActor 查询好友快照和在线状态
	queryMsg := &protocol.FriendListQueryMsg{
		RequesterId:        roleId,
		RequesterSessionId: sessionId,
		FriendIds:          friendList,
	}
	msgData, err := proto.Marshal(queryMsg)
	if err != nil {
		log.Errorf("handleQueryFriendList: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdFriendListQuery), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleQueryFriendList: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// handleRemoveFriend 处理删除好友
func handleRemoveFriend(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	_, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleRemoveFriend: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SRemoveFriendReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleRemoveFriend: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	friendSys := GetFriendSys(ctx)
	if friendSys == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRemoveFriendResult), &protocol.S2CRemoveFriendResultReq{
			Success:  false,
			Message:  "好友系统未初始化",
			FriendId: req.FriendId,
		})
	}

	// 移除好友
	success := friendSys.RemoveFriend(req.FriendId)
	message := "删除成功"
	if !success {
		message = "未找到该好友"
	}
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRemoveFriendResult), &protocol.S2CRemoveFriendResultReq{
		Success:  success,
		Message:  message,
		FriendId: req.FriendId,
	})
}

// handleAddToBlacklist 处理添加到黑名单
func handleAddToBlacklist(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleAddToBlacklist: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SAddToBlacklistReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleAddToBlacklist: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()

	// 验证目标ID
	if req.TargetId == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CAddToBlacklistResult), &protocol.S2CAddToBlacklistResultReq{
			Success: false,
			Message: "目标角色ID无效",
		})
	}

	// 不能拉黑自己
	if req.TargetId == roleId {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CAddToBlacklistResult), &protocol.S2CAddToBlacklistResultReq{
			Success: false,
			Message: "不能拉黑自己",
		})
	}

	// 添加到黑名单
	err = database.AddToBlacklist(req.TargetId, roleId, req.Reason)
	if err != nil {
		log.Errorf("handleAddToBlacklist: add to blacklist failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CAddToBlacklistResult), &protocol.S2CAddToBlacklistResultReq{
			Success: false,
			Message: "添加到黑名单失败",
		})
	}

	log.Infof("Role %d added %d to blacklist, reason: %s", roleId, req.TargetId, req.Reason)

	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CAddToBlacklistResult), &protocol.S2CAddToBlacklistResultReq{
		Success: true,
		Message: "已添加到黑名单",
	})
}

// handleRemoveFromBlacklist 处理从黑名单移除
func handleRemoveFromBlacklist(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleRemoveFromBlacklist: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SRemoveFromBlacklistReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleRemoveFromBlacklist: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()

	// 验证目标ID
	if req.TargetId == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRemoveFromBlacklistResult), &protocol.S2CRemoveFromBlacklistResultReq{
			Success: false,
			Message: "目标角色ID无效",
		})
	}

	// 从黑名单移除
	err = database.RemoveFromBlacklist(req.TargetId, roleId)
	if err != nil {
		log.Errorf("handleRemoveFromBlacklist: remove from blacklist failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRemoveFromBlacklistResult), &protocol.S2CRemoveFromBlacklistResultReq{
			Success: false,
			Message: "从黑名单移除失败",
		})
	}

	log.Infof("Role %d removed %d from blacklist", roleId, req.TargetId)

	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRemoveFromBlacklistResult), &protocol.S2CRemoveFromBlacklistResultReq{
		Success: true,
		Message: "已从黑名单移除",
	})
}

// handleQueryBlacklist 处理查询黑名单
func handleQueryBlacklist(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleQueryBlacklist: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	roleId := playerRole.GetPlayerRoleId()

	// 查询黑名单列表
	blacklists, err := database.GetBlacklist(roleId)
	if err != nil {
		log.Errorf("handleQueryBlacklist: get blacklist failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "查询黑名单失败",
		})
	}

	// 构建黑名单ID列表
	blacklistIds := make([]uint64, 0, len(blacklists))
	for _, blacklist := range blacklists {
		blacklistIds = append(blacklistIds, blacklist.RoleId)
	}

	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CBlacklist), &protocol.S2CBlacklistReq{
		BlacklistIds: blacklistIds,
	})
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysFriend), NewFriendSys)
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SAddFriend), handleAddFriend)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SRespondFriendReq), handleRespondFriendReq)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SQueryFriendList), handleQueryFriendList)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SRemoveFriend), handleRemoveFriend)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SAddToBlacklist), handleAddToBlacklist)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SRemoveFromBlacklist), handleRemoveFromBlacklist)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SQueryBlacklist), handleQueryBlacklist)
	})
}
