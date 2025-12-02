package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/core/iface"
	auctiondomain "postapocgame/server/service/gameserver/internel/domain/auction"
)

// AuctionSystemAdapter 拍卖行系统适配器
type AuctionSystemAdapter struct {
	*BaseSystemAdapter
}

func NewAuctionSystemAdapter() *AuctionSystemAdapter {
	return &AuctionSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysAuction)),
	}
}

func (a *AuctionSystemAdapter) OnInit(ctx context.Context) {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("auction sys OnInit get role err:%v", err)
		return
	}
	auctiondomain.EnsureAuctionData(playerRole.GetBinaryData())
}

var _ iface.ISystem = (*AuctionSystemAdapter)(nil)
