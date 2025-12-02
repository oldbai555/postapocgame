package guild

import (
	"context"
	"unicode/utf8"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	guilddomain "postapocgame/server/service/gameserver/internel/domain/guild"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// CreateGuildUseCase 创建公会
type CreateGuildUseCase struct {
	playerRepo repository.PlayerRepository
	publicGate interfaces.PublicActorGateway
}

// NewCreateGuildUseCase 构造函数
func NewCreateGuildUseCase(playerRepo repository.PlayerRepository, publicGate interfaces.PublicActorGateway) *CreateGuildUseCase {
	return &CreateGuildUseCase{
		playerRepo: playerRepo,
		publicGate: publicGate,
	}
}

// Execute 执行创建逻辑
func (uc *CreateGuildUseCase) Execute(ctx context.Context, roleID uint64, roleName, guildName string) error {
	if roleID == 0 {
		return customerr.NewError("未登录")
	}
	nameLen := utf8.RuneCountInString(guildName)
	if nameLen == 0 || nameLen > 20 {
		return customerr.NewError("公会名称长度必须在1-20个字符之间")
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
		return customerr.NewError("您已经加入公会，无法创建新公会")
	}

	req := &protocol.CreateGuildMsg{
		CreatorId:   roleID,
		GuildName:   guildName,
		CreatorName: roleName,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		return customerr.Wrap(err)
	}
	msg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdCreateGuild), data)
	return customerr.Wrap(uc.publicGate.SendMessageAsync(ctx, "global", msg))
}
