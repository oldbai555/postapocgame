package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	guilddomain "postapocgame/server/service/gameserver/internel/domain/guild"
	"postapocgame/server/service/gameserver/internel/domain/repository"
)

// GuildSystemAdapter 公会系统适配器
type GuildSystemAdapter struct {
	*BaseSystemAdapter
	playerRepo repository.PlayerRepository
}

func NewGuildSystemAdapter() *GuildSystemAdapter {
	return &GuildSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysGuild)),
		playerRepo:        di.GetContainer().PlayerGateway(),
	}
}

func (a *GuildSystemAdapter) OnInit(ctx context.Context) {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("guild sys OnInit get role err:%v", err)
		return
	}
	guilddomain.EnsureGuildData(playerRole.GetBinaryData())
}

var _ iface.ISystem = (*GuildSystemAdapter)(nil)
