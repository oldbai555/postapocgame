package guild

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	guilddomain "postapocgame/server/service/gameserver/internel/domain/guild"
	"postapocgame/server/service/gameserver/internel/domain/repository"
)

// QueryGuildInfoUseCase 查询公会信息
type QueryGuildInfoUseCase struct {
	playerRepo repository.PlayerRepository
}

func NewQueryGuildInfoUseCase(playerRepo repository.PlayerRepository) *QueryGuildInfoUseCase {
	return &QueryGuildInfoUseCase{playerRepo: playerRepo}
}

func (uc *QueryGuildInfoUseCase) Execute(ctx context.Context, roleID uint64) (*protocol.GuildData, error) {
	if roleID == 0 {
		return nil, customerr.NewError("未登录")
	}
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return nil, customerr.Wrap(err)
	}
	siGuild := guilddomain.EnsureGuildData(binaryData)
	if siGuild == nil {
		return nil, customerr.NewError("公会数据异常")
	}
	// 仅返回 GuildId，其他展示字段交由 PublicActor 补全
	return &protocol.GuildData{
		GuildId: siGuild.GuildId,
	}, nil
}
