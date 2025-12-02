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

// JoinGuildUseCase 申请加入公会
type JoinGuildUseCase struct {
	playerRepo repository.PlayerRepository
	publicGate interfaces.PublicActorGateway
}

func NewJoinGuildUseCase(playerRepo repository.PlayerRepository, publicGate interfaces.PublicActorGateway) *JoinGuildUseCase {
	return &JoinGuildUseCase{
		playerRepo: playerRepo,
		publicGate: publicGate,
	}
}

func (uc *JoinGuildUseCase) Execute(ctx context.Context, roleID uint64, roleName string, guildID uint64) error {
	if roleID == 0 {
		return customerr.NewError("未登录")
	}
	if guildID == 0 {
		return customerr.NewError("公会ID无效")
	}

	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return customerr.Wrap(err)
	}
	guildData := guilddomain.EnsureGuildData(binaryData)
	if guildData == nil {
		return customerr.NewError("公会数据异常")
	}
	if guildData.GuildId > 0 {
		return customerr.NewError("您已经加入公会，无法加入新公会")
	}

	req := &protocol.JoinGuildReqMsg{
		RequesterId:   roleID,
		GuildId:       guildID,
		RequesterName: roleName,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		return customerr.Wrap(err)
	}
	msg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdJoinGuildReq), data)
	return customerr.Wrap(uc.publicGate.SendMessageAsync(ctx, "global", msg))
}
