package guild

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	guilddomain "postapocgame/server/service/gameserver/internel/domain/guild"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// LeaveGuildUseCase 离开公会
type LeaveGuildUseCase struct {
	playerRepo repository.PlayerRepository
	publicGate interfaces.PublicActorGateway
}

func NewLeaveGuildUseCase(playerRepo repository.PlayerRepository, publicGate interfaces.PublicActorGateway) *LeaveGuildUseCase {
	return &LeaveGuildUseCase{
		playerRepo: playerRepo,
		publicGate: publicGate,
	}
}

func (uc *LeaveGuildUseCase) Execute(ctx context.Context, roleID uint64) error {
	if roleID == 0 {
		return customerr.NewError("未登录")
	}
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return customerr.Wrap(err)
	}
	guildData := guilddomain.EnsureGuildData(binaryData)
	if guildData == nil {
		return customerr.NewError("公会数据异常")
	}
	if guildData.GuildId == 0 {
		return customerr.NewError("您未加入任何公会")
	}

	req := &protocol.LeaveGuildMsg{
		RoleId:  roleID,
		GuildId: guildData.GuildId,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		return customerr.Wrap(err)
	}
	msg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdLeaveGuild), data)
	if err := uc.publicGate.SendMessageAsync(ctx, "global", msg); err != nil {
		return customerr.Wrap(err)
	}

	// 清理本地数据
	guildData.GuildId = 0
	guildData.Position = 0
	guildData.JoinTime = 0
	return nil
}
