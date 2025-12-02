package publicactor

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/core/gshare"
	"postapocgame/server/service/gameserver/internel/infrastructure/gatewaylink"
)

// 公会数据相关逻辑

// GetGuild 获取公会数据
func (pr *PublicRole) GetGuild(guildId uint64) (*protocol.GuildData, bool) {
	value, ok := pr.guildMap.Load(guildId)
	if !ok {
		return nil, false
	}
	guild, ok := value.(*protocol.GuildData)
	return guild, ok
}

// GetNextGuildId 获取下一个公会ID
func (pr *PublicRole) GetNextGuildId() uint64 {
	pr.guildIdMu.Lock()
	defer pr.guildIdMu.Unlock()
	id := pr.nextGuildId
	pr.nextGuildId++
	return id
}

// ===== 公会相关 handler 注册（无闭包捕获 PublicRole） =====

// RegisterGuildHandlers 注册公会系统相关的消息处理器
func RegisterGuildHandlers(facade gshare.IPublicActorFacade) {
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdCreateGuild), handleCreateGuildMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdJoinGuildReq), handleJoinGuildReqMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdLeaveGuild), handleLeaveGuildMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdGuildJoinApprove), handleGuildJoinApproveMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdGuildJoinReject), handleGuildJoinRejectMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdGuildKickMember), handleGuildKickMemberMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdGuildChangePosition), handleGuildChangePositionMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdGuildUpdateAnnouncement), handleGuildUpdateAnnouncementMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdGuildUpdateName), handleGuildUpdateNameMsg)
}

// 公会 handler 适配
var (
	handleCreateGuildMsg             = withPublicRole(handleCreateGuild)
	handleJoinGuildReqMsg            = withPublicRole(handleJoinGuildReq)
	handleLeaveGuildMsg              = withPublicRole(handleLeaveGuild)
	handleGuildJoinApproveMsg        = withPublicRole(handleGuildJoinApprove)
	handleGuildJoinRejectMsg         = withPublicRole(handleGuildJoinReject)
	handleGuildKickMemberMsg         = withPublicRole(handleGuildKickMember)
	handleGuildChangePositionMsg     = withPublicRole(handleGuildChangePosition)
	handleGuildUpdateAnnouncementMsg = withPublicRole(handleGuildUpdateAnnouncement)
	handleGuildUpdateNameMsg         = withPublicRole(handleGuildUpdateName)
)

// ===== 公会业务 handler（从 message_handler.go 迁移）=====

// handleCreateGuild 处理创建公会
func handleCreateGuild(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	createMsg := &protocol.CreateGuildMsg{}
	if err := proto.Unmarshal(data, createMsg); err != nil {
		log.Errorf("Failed to unmarshal CreateGuildMsg: %v", err)
		return
	}

	// 验证公会名称
	if len(createMsg.GuildName) == 0 || len(createMsg.GuildName) > 20 {
		log.Warnf("handleCreateGuild: invalid guild name length")
		return
	}

	// 检查公会名称是否已存在
	nameExists := false
	publicRole.guildMap.Range(func(key, value interface{}) bool {
		guild, ok := value.(*protocol.GuildData)
		if ok && guild.GuildName == createMsg.GuildName {
			nameExists = true
			return false // 停止遍历
		}
		return true
	})

	if nameExists {
		log.Warnf("handleCreateGuild: guild name %s already exists", createMsg.GuildName)
		// 发送失败响应
		creatorSessionId, ok := publicRole.GetSessionId(createMsg.CreatorId)
		if ok {
			respMsg := &protocol.S2CCreateGuildResultReq{
				Success: false,
				Message: "公会名称已存在",
			}
			respData, err := proto.Marshal(respMsg)
			if err == nil {
				gatewaylink.SendToSession(creatorSessionId, uint16(protocol.S2CProtocol_S2CCreateGuildResult), respData)
			}
		}
		return
	}

	// 创建公会
	guildId := publicRole.GetNextGuildId()
	guildData := &protocol.GuildData{
		GuildId:      guildId,
		GuildName:    createMsg.GuildName,
		CreatorId:    createMsg.CreatorId,
		Level:        1,
		CreateTime:   servertime.UnixMilli(),
		Members:      make([]*protocol.GuildMember, 0),
		Announcement: "",
	}

	// 添加创建者为会长
	guildData.Members = append(guildData.Members, &protocol.GuildMember{
		RoleId:   createMsg.CreatorId,
		Position: uint32(protocol.GuildPosition_GuildPositionLeader),
		JoinTime: servertime.UnixMilli(),
	})

	// 存储公会数据
	publicRole.SetGuild(guildId, guildData)

	// 发送响应给创建者
	creatorSessionId, ok := publicRole.GetSessionId(createMsg.CreatorId)
	if ok {
		respMsg := &protocol.S2CCreateGuildResultReq{
			Success:   true,
			Message:   "公会创建成功",
			GuildData: guildData,
		}
		respData, err := proto.Marshal(respMsg)
		if err == nil {
			gatewaylink.SendToSession(creatorSessionId, uint16(protocol.S2CProtocol_S2CCreateGuildResult), respData)
		}
	}

	log.Debugf("handleCreateGuild: created guild %d with name %s by role %d", guildId, createMsg.GuildName, createMsg.CreatorId)
}

// handleJoinGuildReq 处理加入公会请求
func handleJoinGuildReq(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	joinMsg := &protocol.JoinGuildReqMsg{}
	if err := proto.Unmarshal(data, joinMsg); err != nil {
		log.Errorf("Failed to unmarshal JoinGuildReqMsg: %v", err)
		return
	}

	// 查找公会
	guild, ok := publicRole.GetGuild(joinMsg.GuildId)
	if !ok {
		log.Warnf("handleJoinGuildReq: guild %d not found", joinMsg.GuildId)
		return
	}

	// 检查是否已经是成员
	for _, member := range guild.Members {
		if member.RoleId == joinMsg.RequesterId {
			log.Warnf("handleJoinGuildReq: role %d already in guild %d", joinMsg.RequesterId, joinMsg.GuildId)
			return
		}
	}

	// 添加成员
	guild.Members = append(guild.Members, &protocol.GuildMember{
		RoleId:   joinMsg.RequesterId,
		Position: uint32(protocol.GuildPosition_GuildPositionMember),
		JoinTime: servertime.UnixMilli(),
	})

	// 更新公会数据
	publicRole.SetGuild(joinMsg.GuildId, guild)

	log.Debugf("handleJoinGuildReq: role %d joined guild %d", joinMsg.RequesterId, joinMsg.GuildId)
}

// handleLeaveGuild 处理离开公会
func handleLeaveGuild(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	leaveMsg := &protocol.LeaveGuildMsg{}
	if err := proto.Unmarshal(data, leaveMsg); err != nil {
		log.Errorf("Failed to unmarshal LeaveGuildMsg: %v", err)
		return
	}

	// 查找公会
	guild, ok := publicRole.GetGuild(leaveMsg.GuildId)
	if !ok {
		log.Warnf("handleLeaveGuild: guild %d not found", leaveMsg.GuildId)
		return
	}

	// 移除成员
	newMembers := make([]*protocol.GuildMember, 0)
	for _, member := range guild.Members {
		if member.RoleId != leaveMsg.RoleId {
			newMembers = append(newMembers, member)
		}
	}
	guild.Members = newMembers

	// 如果公会为空，删除公会
	if len(guild.Members) == 0 {
		publicRole.DeleteGuild(leaveMsg.GuildId)
		log.Debugf("handleLeaveGuild: guild %d deleted (no members)", leaveMsg.GuildId)
	} else {
		// 更新公会数据
		publicRole.SetGuild(leaveMsg.GuildId, guild)
		log.Debugf("handleLeaveGuild: role %d left guild %d", leaveMsg.RoleId, leaveMsg.GuildId)
	}
}

// handleGuildJoinApprove 处理审批加入公会申请
func handleGuildJoinApprove(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	approveMsg := &protocol.GuildJoinApproveMsg{}
	if err := proto.Unmarshal(data, approveMsg); err != nil {
		log.Errorf("Failed to unmarshal GuildJoinApproveMsg: %v", err)
		return
	}

	// 查找公会
	guild, ok := publicRole.GetGuild(approveMsg.GuildId)
	if !ok {
		log.Warnf("handleGuildJoinApprove: guild %d not found", approveMsg.GuildId)
		return
	}

	// 检查权限
	if !CheckGuildPermission(guild, approveMsg.ApproverId, PermissionGuildApproveJoin) {
		log.Warnf("handleGuildJoinApprove: approver %d has no permission", approveMsg.ApproverId)
		approverSessionId, ok := publicRole.GetSessionId(approveMsg.ApproverId)
		if ok {
			gatewaylink.SendToSessionProto(approverSessionId, uint16(protocol.S2CProtocol_S2CGuildJoinApproveResult), &protocol.S2CGuildJoinApproveResultReq{
				Success: false,
				Message: "没有审批权限",
			})
		}
		return
	}

	// 检查申请是否存在
	applications := publicRole.GetGuildApplications(approveMsg.GuildId)
	found := false
	for _, app := range applications {
		if app.ApplicantId == approveMsg.ApplicantId {
			found = true
			break
		}
	}
	if !found {
		log.Warnf("handleGuildJoinApprove: application not found for applicant %d", approveMsg.ApplicantId)
		approverSessionId, ok := publicRole.GetSessionId(approveMsg.ApproverId)
		if ok {
			gatewaylink.SendToSessionProto(approverSessionId, uint16(protocol.S2CProtocol_S2CGuildJoinApproveResult), &protocol.S2CGuildJoinApproveResultReq{
				Success: false,
				Message: "申请不存在",
			})
		}
		return
	}

	// 检查申请人是否已经在公会中
	for _, member := range guild.Members {
		if member.RoleId == approveMsg.ApplicantId {
			log.Warnf("handleGuildJoinApprove: applicant %d already in guild", approveMsg.ApplicantId)
			// 移除申请
			publicRole.RemoveGuildApplication(approveMsg.GuildId, approveMsg.ApplicantId)
			return
		}
	}

	// 添加成员
	guild.Members = append(guild.Members, &protocol.GuildMember{
		RoleId:   approveMsg.ApplicantId,
		Position: uint32(protocol.GuildPosition_GuildPositionMember),
		JoinTime: servertime.UnixMilli(),
	})

	// 更新公会数据
	publicRole.SetGuild(approveMsg.GuildId, guild)

	// 移除申请
	publicRole.RemoveGuildApplication(approveMsg.GuildId, approveMsg.ApplicantId)

	// 通知审批者
	approverSessionId, ok := publicRole.GetSessionId(approveMsg.ApproverId)
	if ok {
		gatewaylink.SendToSessionProto(approverSessionId, uint16(protocol.S2CProtocol_S2CGuildJoinApproveResult), &protocol.S2CGuildJoinApproveResultReq{
			Success: true,
			Message: "审批成功",
		})
	}

	// 通知申请人
	applicantSessionId, ok := publicRole.GetSessionId(approveMsg.ApplicantId)
	if ok {
		gatewaylink.SendToSessionProto(applicantSessionId, uint16(protocol.S2CProtocol_S2CJoinGuildResult), &protocol.S2CJoinGuildResultReq{
			Success:   true,
			Message:   "加入公会成功",
			GuildData: guild,
		})
	}

	log.Debugf("handleGuildJoinApprove: applicant %d approved to join guild %d", approveMsg.ApplicantId, approveMsg.GuildId)
}

// handleGuildJoinReject 处理拒绝加入公会申请
func handleGuildJoinReject(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	rejectMsg := &protocol.GuildJoinRejectMsg{}
	if err := proto.Unmarshal(data, rejectMsg); err != nil {
		log.Errorf("Failed to unmarshal GuildJoinRejectMsg: %v", err)
		return
	}

	// 查找公会
	guild, ok := publicRole.GetGuild(rejectMsg.GuildId)
	if !ok {
		log.Warnf("handleGuildJoinReject: guild %d not found", rejectMsg.GuildId)
		return
	}

	// 检查权限
	if !CheckGuildPermission(guild, rejectMsg.ApproverId, PermissionGuildApproveJoin) {
		log.Warnf("handleGuildJoinReject: approver %d has no permission", rejectMsg.ApproverId)
		approverSessionId, ok := publicRole.GetSessionId(rejectMsg.ApproverId)
		if ok {
			gatewaylink.SendToSessionProto(approverSessionId, uint16(protocol.S2CProtocol_S2CGuildJoinRejectResult), &protocol.S2CGuildJoinRejectResultReq{
				Success: false,
				Message: "没有审批权限",
			})
		}
		return
	}

	// 移除申请
	publicRole.RemoveGuildApplication(rejectMsg.GuildId, rejectMsg.ApplicantId)

	// 通知审批者
	approverSessionId, ok := publicRole.GetSessionId(rejectMsg.ApproverId)
	if ok {
		gatewaylink.SendToSessionProto(approverSessionId, uint16(protocol.S2CProtocol_S2CGuildJoinRejectResult), &protocol.S2CGuildJoinRejectResultReq{
			Success: true,
			Message: "已拒绝申请",
		})
	}

	// 通知申请人
	applicantSessionId, ok := publicRole.GetSessionId(rejectMsg.ApplicantId)
	if ok {
		gatewaylink.SendToSessionProto(applicantSessionId, uint16(protocol.S2CProtocol_S2CJoinGuildResult), &protocol.S2CJoinGuildResultReq{
			Success: false,
			Message: "加入公会被拒绝",
		})
	}
}

// handleGuildKickMember 处理踢出成员
func handleGuildKickMember(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	kickMsg := &protocol.GuildKickMemberMsg{}
	if err := proto.Unmarshal(data, kickMsg); err != nil {
		log.Errorf("Failed to unmarshal GuildKickMemberMsg: %v", err)
		return
	}

	// 查找公会
	guild, ok := publicRole.GetGuild(kickMsg.GuildId)
	if !ok {
		log.Warnf("handleGuildKickMember: guild %d not found", kickMsg.GuildId)
		return
	}

	// 检查权限
	if !CheckGuildPermission(guild, kickMsg.OperatorId, PermissionGuildKickMember) {
		log.Warnf("handleGuildKickMember: operator %d has no permission", kickMsg.OperatorId)
		operatorSessionId, ok := publicRole.GetSessionId(kickMsg.OperatorId)
		if ok {
			gatewaylink.SendToSessionProto(operatorSessionId, uint16(protocol.S2CProtocol_S2CGuildKickMemberResult), &protocol.S2CGuildKickMemberResultReq{
				Success: false,
				Message: "没有踢出成员权限",
			})
		}
		return
	}

	// 检查目标成员是否存在
	targetPos, ok := GetGuildMemberPosition(guild, kickMsg.TargetId)
	if !ok {
		log.Warnf("handleGuildKickMember: target %d not in guild", kickMsg.TargetId)
		return
	}

	// 会长不能踢出自己
	if kickMsg.OperatorId == kickMsg.TargetId {
		log.Warnf("handleGuildKickMember: cannot kick self")
		return
	}

	// 副会长不能踢出会长和副会长
	operatorPos, _ := GetGuildMemberPosition(guild, kickMsg.OperatorId)
	if protocol.GuildPosition(operatorPos) == protocol.GuildPosition_GuildPositionViceLeader {
		if protocol.GuildPosition(targetPos) == protocol.GuildPosition_GuildPositionLeader ||
			protocol.GuildPosition(targetPos) == protocol.GuildPosition_GuildPositionViceLeader {
			log.Warnf("handleGuildKickMember: vice leader cannot kick leader or vice leader")
			return
		}
	}

	// 移除成员
	newMembers := make([]*protocol.GuildMember, 0)
	for _, member := range guild.Members {
		if member.RoleId != kickMsg.TargetId {
			newMembers = append(newMembers, member)
		}
	}
	guild.Members = newMembers

	// 如果公会为空，删除公会
	if len(guild.Members) == 0 {
		publicRole.DeleteGuild(kickMsg.GuildId)
		log.Debugf("handleGuildKickMember: guild %d deleted (no members)", kickMsg.GuildId)
	} else {
		// 更新公会数据
		publicRole.SetGuild(kickMsg.GuildId, guild)
	}

	// 通知操作者
	operatorSessionId, ok := publicRole.GetSessionId(kickMsg.OperatorId)
	if ok {
		gatewaylink.SendToSessionProto(operatorSessionId, uint16(protocol.S2CProtocol_S2CGuildKickMemberResult), &protocol.S2CGuildKickMemberResultReq{
			Success: true,
			Message: "已踢出成员",
		})
	}

	// 通知被踢出的成员
	targetSessionId, ok := publicRole.GetSessionId(kickMsg.TargetId)
	if ok {
		gatewaylink.SendToSessionProto(targetSessionId, uint16(protocol.S2CProtocol_S2CLeaveGuildResult), &protocol.S2CLeaveGuildResultReq{
			Success: true,
			Message: "你已被踢出公会",
		})
	}

	log.Debugf("handleGuildKickMember: member %d kicked from guild %d by %d", kickMsg.TargetId, kickMsg.GuildId, kickMsg.OperatorId)
}

// handleGuildChangePosition 处理修改职位
func handleGuildChangePosition(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	changeMsg := &protocol.GuildChangePositionMsg{}
	if err := proto.Unmarshal(data, changeMsg); err != nil {
		log.Errorf("Failed to unmarshal GuildChangePositionMsg: %v", err)
		return
	}

	// 查找公会
	guild, ok := publicRole.GetGuild(changeMsg.GuildId)
	if !ok {
		log.Warnf("handleGuildChangePosition: guild %d not found", changeMsg.GuildId)
		return
	}

	// 检查权限
	if !CanChangePosition(guild, changeMsg.OperatorId, changeMsg.TargetId, changeMsg.NewPosition) {
		log.Warnf("handleGuildChangePosition: operator %d cannot change position", changeMsg.OperatorId)
		operatorSessionId, ok := publicRole.GetSessionId(changeMsg.OperatorId)
		if ok {
			gatewaylink.SendToSessionProto(operatorSessionId, uint16(protocol.S2CProtocol_S2CGuildChangePositionResult), &protocol.S2CGuildChangePositionResultReq{
				Success: false,
				Message: "没有修改职位权限",
			})
		}
		return
	}

	// 更新职位
	for _, member := range guild.Members {
		if member.RoleId == changeMsg.TargetId {
			member.Position = changeMsg.NewPosition
			break
		}
	}

	// 更新公会数据
	publicRole.SetGuild(changeMsg.GuildId, guild)

	// 通知操作者
	operatorSessionId, ok := publicRole.GetSessionId(changeMsg.OperatorId)
	if ok {
		gatewaylink.SendToSessionProto(operatorSessionId, uint16(protocol.S2CProtocol_S2CGuildChangePositionResult), &protocol.S2CGuildChangePositionResultReq{
			Success: true,
			Message: "职位修改成功",
		})
	}

	// 通知目标成员
	targetSessionId, ok := publicRole.GetSessionId(changeMsg.TargetId)
	if ok {
		gatewaylink.SendToSessionProto(targetSessionId, uint16(protocol.S2CProtocol_S2CGuildInfo), &protocol.S2CGuildInfoReq{
			GuildData: guild,
		})
	}

	log.Debugf("handleGuildChangePosition: member %d position changed to %d in guild %d", changeMsg.TargetId, changeMsg.NewPosition, changeMsg.GuildId)
}

// handleGuildUpdateAnnouncement 处理修改公会宣言
func handleGuildUpdateAnnouncement(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	updateMsg := &protocol.GuildUpdateAnnouncementMsg{}
	if err := proto.Unmarshal(data, updateMsg); err != nil {
		log.Errorf("Failed to unmarshal GuildUpdateAnnouncementMsg: %v", err)
		return
	}

	// 验证宣言长度
	if len(updateMsg.Announcement) > 200 {
		log.Warnf("handleGuildUpdateAnnouncement: announcement too long")
		operatorSessionId, ok := publicRole.GetSessionId(updateMsg.OperatorId)
		if ok {
			gatewaylink.SendToSessionProto(operatorSessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
				Code: -1,
				Msg:  "宣言长度不能超过200字符",
			})
		}
		return
	}

	// 查找公会
	guild, ok := publicRole.GetGuild(updateMsg.GuildId)
	if !ok {
		log.Warnf("handleGuildUpdateAnnouncement: guild %d not found", updateMsg.GuildId)
		return
	}

	// 检查权限
	if !CheckGuildPermission(guild, updateMsg.OperatorId, PermissionGuildUpdateAnn) {
		log.Warnf("handleGuildUpdateAnnouncement: operator %d has no permission", updateMsg.OperatorId)
		operatorSessionId, ok := publicRole.GetSessionId(updateMsg.OperatorId)
		if ok {
			gatewaylink.SendToSessionProto(operatorSessionId, uint16(protocol.S2CProtocol_S2CGuildUpdateAnnouncementResult), &protocol.S2CGuildUpdateAnnouncementResultReq{
				Success: false,
				Message: "没有修改宣言权限",
			})
		}
		return
	}

	// 更新宣言
	guild.Announcement = updateMsg.Announcement

	// 更新公会数据
	publicRole.SetGuild(updateMsg.GuildId, guild)

	// 通知操作者
	operatorSessionId, ok := publicRole.GetSessionId(updateMsg.OperatorId)
	if ok {
		gatewaylink.SendToSessionProto(operatorSessionId, uint16(protocol.S2CProtocol_S2CGuildUpdateAnnouncementResult), &protocol.S2CGuildUpdateAnnouncementResultReq{
			Success: true,
			Message: "宣言修改成功",
		})
	}

	// 通知所有成员
	for _, member := range guild.Members {
		memberSessionId, ok := publicRole.GetSessionId(member.RoleId)
		if ok {
			gatewaylink.SendToSessionProto(memberSessionId, uint16(protocol.S2CProtocol_S2CGuildInfo), &protocol.S2CGuildInfoReq{
				GuildData: guild,
			})
		}
	}

	log.Debugf("handleGuildUpdateAnnouncement: guild %d announcement updated by %d", updateMsg.GuildId, updateMsg.OperatorId)
}

// handleGuildUpdateName 处理修改公会名称
func handleGuildUpdateName(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	updateMsg := &protocol.GuildUpdateNameMsg{}
	if err := proto.Unmarshal(data, updateMsg); err != nil {
		log.Errorf("Failed to unmarshal GuildUpdateNameMsg: %v", err)
		return
	}

	// 验证名称长度
	if len(updateMsg.NewName) == 0 || len(updateMsg.NewName) > 20 {
		log.Warnf("handleGuildUpdateName: invalid name length")
		operatorSessionId, ok := publicRole.GetSessionId(updateMsg.OperatorId)
		if ok {
			gatewaylink.SendToSessionProto(operatorSessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
				Code: -1,
				Msg:  "公会名称长度必须在1-20字符之间",
			})
		}
		return
	}

	// 查找公会
	guild, ok := publicRole.GetGuild(updateMsg.GuildId)
	if !ok {
		log.Warnf("handleGuildUpdateName: guild %d not found", updateMsg.GuildId)
		return
	}

	// 检查权限
	if !CheckGuildPermission(guild, updateMsg.OperatorId, PermissionGuildUpdateName) {
		log.Warnf("handleGuildUpdateName: operator %d has no permission", updateMsg.OperatorId)
		operatorSessionId, ok := publicRole.GetSessionId(updateMsg.OperatorId)
		if ok {
			gatewaylink.SendToSessionProto(operatorSessionId, uint16(protocol.S2CProtocol_S2CGuildUpdateNameResult), &protocol.S2CGuildUpdateNameResultReq{
				Success: false,
				Message: "没有修改名称权限",
			})
		}
		return
	}

	// 检查名称是否已存在
	nameExists := false
	publicRole.guildMap.Range(func(key, value interface{}) bool {
		g, ok := value.(*protocol.GuildData)
		if ok && g.GuildId != updateMsg.GuildId && g.GuildName == updateMsg.NewName {
			nameExists = true
			return false
		}
		return true
	})

	if nameExists {
		log.Warnf("handleGuildUpdateName: guild name %s already exists", updateMsg.NewName)
		operatorSessionId, ok := publicRole.GetSessionId(updateMsg.OperatorId)
		if ok {
			gatewaylink.SendToSessionProto(operatorSessionId, uint16(protocol.S2CProtocol_S2CGuildUpdateNameResult), &protocol.S2CGuildUpdateNameResultReq{
				Success: false,
				Message: "公会名称已存在",
			})
		}
		return
	}

	// 更新名称
	guild.GuildName = updateMsg.NewName

	// 更新公会数据
	publicRole.SetGuild(updateMsg.GuildId, guild)

	// 通知操作者
	operatorSessionId, ok := publicRole.GetSessionId(updateMsg.OperatorId)
	if ok {
		gatewaylink.SendToSessionProto(operatorSessionId, uint16(protocol.S2CProtocol_S2CGuildUpdateNameResult), &protocol.S2CGuildUpdateNameResultReq{
			Success: true,
			Message: "名称修改成功",
		})
	}

	// 通知所有成员
	for _, member := range guild.Members {
		memberSessionId, ok := publicRole.GetSessionId(member.RoleId)
		if ok {
			gatewaylink.SendToSessionProto(memberSessionId, uint16(protocol.S2CProtocol_S2CGuildInfo), &protocol.S2CGuildInfoReq{
				GuildData: guild,
			})
		}
	}

	log.Debugf("handleGuildUpdateName: guild %d name updated to %s by %d", updateMsg.GuildId, updateMsg.NewName, updateMsg.OperatorId)
}
