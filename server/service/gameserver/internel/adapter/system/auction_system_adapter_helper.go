package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

func GetAuctionSys(ctx context.Context) *AuctionSystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysAuction))
	if system == nil {
		return nil
	}
	auctionSys, ok := system.(*AuctionSystemAdapter)
	if !ok || !auctionSys.IsOpened() {
		return nil
	}
	return auctionSys
}
