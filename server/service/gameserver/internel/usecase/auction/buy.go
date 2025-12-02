package auction

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// BuyUseCase 拍卖购买
type BuyUseCase struct {
	publicGate interfaces.PublicActorGateway
}

func NewBuyUseCase(publicGate interfaces.PublicActorGateway) *BuyUseCase {
	return &BuyUseCase{publicGate: publicGate}
}

func (uc *BuyUseCase) Execute(ctx context.Context, roleID uint64, auctionID uint64) error {
	if roleID == 0 {
		return customerr.NewError("未登录")
	}
	if auctionID == 0 {
		return customerr.NewError("拍卖品ID无效")
	}
	msg := &protocol.AuctionBuyMsg{
		BuyerId:   roleID,
		AuctionId: auctionID,
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return customerr.Wrap(err)
	}
	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAuctionBuy), data)
	return customerr.Wrap(uc.publicGate.SendMessageAsync(ctx, "global", actorMsg))
}
