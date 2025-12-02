package publicactor

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/core/gshare"
	"postapocgame/server/service/gameserver/internel/infrastructure/gatewaylink"
)

// 好友系统相关 handler 注册

// RegisterFriendHandlers 注册好友系统相关的消息处理器
func RegisterFriendHandlers(facade gshare.IPublicActorFacade) {
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdAddFriendReq), handleAddFriendReqMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdAddFriendResp), handleAddFriendRespMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdFriendListQuery), handleFriendListQueryMsg)
}

// 好友 handler 适配
var (
	handleAddFriendReqMsg    = withPublicRole(handleAddFriendReq)
	handleAddFriendRespMsg   = withPublicRole(handleAddFriendResp)
	handleFriendListQueryMsg = withPublicRole(handleFriendListQuery)
)

// ===== 好友业务 handler（从 message_handler.go 迁移）=====

// handleAddFriendReq 处理好友申请
func handleAddFriendReq(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	addFriendMsg := &protocol.AddFriendReqMsg{}
	if err := proto.Unmarshal(data, addFriendMsg); err != nil {
		log.Errorf("Failed to unmarshal AddFriendReqMsg: %v", err)
		return
	}

	// 检查黑名单
	isBlocked, err := database.IsInBlacklist(addFriendMsg.RequesterId, addFriendMsg.TargetId)
	if err == nil && isBlocked {
		log.Debugf("handleAddFriendReq: requester %d is blocked by target %d", addFriendMsg.RequesterId, addFriendMsg.TargetId)
		// 不发送好友申请，静默处理
		return
	}

	// 验证目标角色是否存在（检查是否在线或从数据库查询）
	if !publicRole.IsOnline(addFriendMsg.TargetId) {
		log.Debugf("handleAddFriendReq: target %d is not online", addFriendMsg.TargetId)
		// 目标不在线，可以选择存储申请（这里先忽略，后续可以完善）
		return
	}

	// 获取目标玩家的 SessionId
	targetSessionId, ok := publicRole.GetSessionId(addFriendMsg.TargetId)
	if !ok {
		log.Warnf("handleAddFriendReq: target %d session not found", addFriendMsg.TargetId)
		return
	}

	// 发送好友申请通知给目标玩家
	notifyMsg := &protocol.S2CFriendRequestNotifyReq{
		RequesterId:   addFriendMsg.RequesterId,
		RequesterName: addFriendMsg.RequesterName,
	}
	notifyData, err := proto.Marshal(notifyMsg)
	if err != nil {
		log.Errorf("Failed to marshal FriendRequestNotify: %v", err)
		return
	}

	err = gatewaylink.SendToSession(targetSessionId, uint16(protocol.S2CProtocol_S2CFriendRequestNotify), notifyData)
	if err != nil {
		log.Warnf("Failed to send friend request notify to session %s: %v", targetSessionId, err)
	}

	log.Debugf("handleAddFriendReq: sent friend request from %d to %d", addFriendMsg.RequesterId, addFriendMsg.TargetId)
}

// handleAddFriendResp 处理好友申请响应（同意/拒绝）
func handleAddFriendResp(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	respMsg := &protocol.AddFriendRespMsg{}
	if err := proto.Unmarshal(data, respMsg); err != nil {
		log.Errorf("Failed to unmarshal AddFriendRespMsg: %v", err)
		return
	}

	// 如果同意，通知申请者好友已添加，并让申请者自动添加目标为好友
	if respMsg.Accepted {
		// 获取申请者的SessionId
		requesterSessionId, ok := publicRole.GetSessionId(respMsg.RequesterId)
		if ok {
			// 发送通知给申请者客户端
			notifyMsg := &protocol.S2CAddFriendResultReq{
				Success:  true,
				Message:  "好友申请已同意",
				TargetId: respMsg.TargetId,
			}
			notifyData, err := proto.Marshal(notifyMsg)
			if err == nil {
				gatewaylink.SendToSession(requesterSessionId, uint16(protocol.S2CProtocol_S2CAddFriendResult), notifyData)
			}
		}
		log.Debugf("handleAddFriendResp: friend request accepted, requester %d, target %d", respMsg.RequesterId, respMsg.TargetId)
	} else {
		// 拒绝，通知申请者
		requesterSessionId, ok := publicRole.GetSessionId(respMsg.RequesterId)
		if ok {
			notifyMsg := &protocol.S2CAddFriendResultReq{
				Success:  false,
				Message:  "好友申请被拒绝",
				TargetId: respMsg.TargetId,
			}
			notifyData, err := proto.Marshal(notifyMsg)
			if err == nil {
				gatewaylink.SendToSession(requesterSessionId, uint16(protocol.S2CProtocol_S2CAddFriendResult), notifyData)
			}
		}
		log.Debugf("handleAddFriendResp: friend request rejected, requester %d, target %d", respMsg.RequesterId, respMsg.TargetId)
	}
}

// handleFriendListQuery 处理好友列表查询
func handleFriendListQuery(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	queryMsg := &protocol.FriendListQueryMsg{}
	if err := proto.Unmarshal(data, queryMsg); err != nil {
		log.Errorf("Failed to unmarshal FriendListQueryMsg: %v", err)
		return
	}

	// 获取好友快照和在线状态
	var snapshots []*protocol.PlayerRankSnapshot
	onlineStatus := make(map[uint64]bool)

	for _, friendId := range queryMsg.FriendIds {
		// 检查在线状态
		isOnline := publicRole.IsOnline(friendId)
		onlineStatus[friendId] = isOnline

		// 获取快照
		snapshot, ok := publicRole.GetRankSnapshot(friendId)
		if ok {
			snapshots = append(snapshots, snapshot)
		} else {
			// 快照不存在，创建基础快照（后续可以从数据库加载）
			snapshots = append(snapshots, &protocol.PlayerRankSnapshot{
				RoleId: friendId,
			})
		}
	}

	// 构建响应消息（转换为 S2C 协议格式）
	friendListResp := &protocol.S2CFriendListReq{
		Friends:      snapshots,
		OnlineStatus: onlineStatus,
	}
	respData, err := proto.Marshal(friendListResp)
	if err != nil {
		log.Errorf("Failed to marshal S2CFriendListReq: %v", err)
		return
	}

	// 发送给请求者
	err = gatewaylink.SendToSession(queryMsg.RequesterSessionId, uint16(protocol.S2CProtocol_S2CFriendList), respData)
	if err != nil {
		log.Warnf("Failed to send friend list to session %s: %v", queryMsg.RequesterSessionId, err)
	}

	log.Debugf("handleFriendListQuery: sent friend list for role %d with %d friends", queryMsg.RequesterId, len(snapshots))
}
