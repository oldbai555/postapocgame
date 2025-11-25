package entitysystem

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
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

// GuildSys 公会系统
type GuildSys struct {
	*BaseSystem
	data *protocol.SiGuildData
}

// NewGuildSys 创建公会系统
func NewGuildSys() iface.ISystem {
	return &GuildSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysGuild)),
	}
}

func (s *GuildSys) OnInit(ctx context.Context) {
	role, err := GetIPlayerRoleByContext(ctx)
	if err != nil || role == nil {
		return
	}
	bd := role.GetBinaryData()
	if bd.GuildData == nil {
		bd.GuildData = &protocol.SiGuildData{
			GuildId:  0,
			Position: 0,
			JoinTime: 0,
		}
	}
	s.data = bd.GuildData
}

// GetGuildId 获取公会ID
func (s *GuildSys) GetGuildId() uint64 {
	if s.data == nil {
		return 0
	}
	return s.data.GuildId
}

// SetGuildId 设置公会ID
func (s *GuildSys) SetGuildId(guildId uint64) {
	if s.data != nil {
		s.data.GuildId = guildId
	}
}

// GetPosition 获取职位
func (s *GuildSys) GetPosition() uint32 {
	if s.data == nil {
		return 0
	}
	return s.data.Position
}

// SetPosition 设置职位
func (s *GuildSys) SetPosition(position uint32) {
	if s.data != nil {
		s.data.Position = position
	}
}

// GetJoinTime 获取加入时间
func (s *GuildSys) GetJoinTime() int64 {
	if s.data == nil {
		return 0
	}
	return s.data.JoinTime
}

// SetJoinTime 设置加入时间
func (s *GuildSys) SetJoinTime(joinTime int64) {
	if s.data != nil {
		s.data.JoinTime = joinTime
	}
}

// GetGuildSys 获取公会系统
func GetGuildSys(ctx context.Context) *GuildSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysGuild))
	if system == nil {
		return nil
	}
	sys := system.(*GuildSys)
	if sys == nil || !sys.IsOpened() {
		return nil
	}
	return sys
}

// handleCreateGuild 处理创建公会
func handleCreateGuild(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleCreateGuild: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SCreateGuildReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleCreateGuild: unmarshal failed: %v", err)
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

	// 验证公会名称
	if len(req.GuildName) == 0 || len(req.GuildName) > 20 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CCreateGuildResult), &protocol.S2CCreateGuildResultReq{
			Success: false,
			Message: "公会名称长度必须在1-20个字符之间",
		})
	}

	// 检查是否已有公会
	guildSys := GetGuildSys(ctx)
	if guildSys != nil && guildSys.GetGuildId() > 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CCreateGuildResult), &protocol.S2CCreateGuildResultReq{
			Success: false,
			Message: "您已经加入公会，无法创建新公会",
		})
	}

	// 发送到 PublicActor 处理
	createMsg := &protocol.CreateGuildMsg{
		CreatorId:   roleId,
		GuildName:   req.GuildName,
		CreatorName: roleInfo.RoleName,
	}
	msgData, err := proto.Marshal(createMsg)
	if err != nil {
		log.Errorf("handleCreateGuild: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdCreateGuild), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleCreateGuild: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	// 这里需要等待 PublicActor 回调，暂时先返回成功（后续完善）
	return nil
}

// handleJoinGuild 处理加入公会
func handleJoinGuild(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleJoinGuild: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SJoinGuildReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleJoinGuild: unmarshal failed: %v", err)
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

	// 检查是否已有公会
	guildSys := GetGuildSys(ctx)
	if guildSys != nil && guildSys.GetGuildId() > 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CJoinGuildResult), &protocol.S2CJoinGuildResultReq{
			Success: false,
			Message: "您已经加入公会，无法加入新公会",
		})
	}

	// 发送到 PublicActor 处理
	joinMsg := &protocol.JoinGuildReqMsg{
		RequesterId:   roleId,
		GuildId:       req.GuildId,
		RequesterName: roleInfo.RoleName,
	}
	msgData, err := proto.Marshal(joinMsg)
	if err != nil {
		log.Errorf("handleJoinGuild: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdJoinGuildReq), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleJoinGuild: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// handleLeaveGuild 处理离开公会
func handleLeaveGuild(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleLeaveGuild: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	roleId := playerRole.GetPlayerRoleId()
	guildSys := GetGuildSys(ctx)
	if guildSys == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CLeaveGuildResult), &protocol.S2CLeaveGuildResultReq{
			Success: false,
			Message: "公会系统未初始化",
		})
	}

	guildId := guildSys.GetGuildId()
	if guildId == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CLeaveGuildResult), &protocol.S2CLeaveGuildResultReq{
			Success: false,
			Message: "您未加入任何公会",
		})
	}

	// 发送到 PublicActor 处理
	leaveMsg := &protocol.LeaveGuildMsg{
		RoleId:  roleId,
		GuildId: guildId,
	}
	msgData, err := proto.Marshal(leaveMsg)
	if err != nil {
		log.Errorf("handleLeaveGuild: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdLeaveGuild), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleLeaveGuild: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	// 清除玩家的公会数据
	guildSys.SetGuildId(0)
	guildSys.SetPosition(0)
	guildSys.SetJoinTime(0)

	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CLeaveGuildResult), &protocol.S2CLeaveGuildResultReq{
		Success: true,
		Message: "离开公会成功",
	})
}

// handleQueryGuildInfo 处理查询公会信息
func handleQueryGuildInfo(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	_, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleQueryGuildInfo: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	guildSys := GetGuildSys(ctx)
	if guildSys == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CGuildInfo), &protocol.S2CGuildInfoReq{
			GuildData: nil,
		})
	}

	guildId := guildSys.GetGuildId()
	if guildId == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CGuildInfo), &protocol.S2CGuildInfoReq{
			GuildData: nil,
		})
	}

	// 需要从 PublicActor 获取公会数据（这里先返回空，后续完善）
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CGuildInfo), &protocol.S2CGuildInfoReq{
		GuildData: nil,
	})
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysGuild), NewGuildSys)
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SCreateGuild), handleCreateGuild)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SJoinGuild), handleJoinGuild)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SLeaveGuild), handleLeaveGuild)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SQueryGuildInfo), handleQueryGuildInfo)
	})
}
