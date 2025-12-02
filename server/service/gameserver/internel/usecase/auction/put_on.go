package auction

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// PutOnUseCase 拍卖上架
type PutOnUseCase struct {
	publicGate interfaces.PublicActorGateway
}

func NewPutOnUseCase(publicGate interfaces.PublicActorGateway) *PutOnUseCase {
	return &PutOnUseCase{publicGate: publicGate}
}

func (uc *PutOnUseCase) Execute(ctx context.Context, sellerID uint64, req *protocol.C2SAuctionPutOnReq) error {
	if sellerID == 0 {
		return customerr.NewError("未登录")
	}
	if req == nil || req.ItemId == 0 || req.Count == 0 || req.Price <= 0 {
		return customerr.NewError("上架参数无效")
	}
	msg := &protocol.AuctionPutOnMsg{
		SellerId: sellerID,
		ItemId:   req.ItemId,
		Count:    req.Count,
		Price:    req.Price,
		Duration: req.Duration,
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return customerr.Wrap(err)
	}
	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAuctionPutOn), data)
	return customerr.Wrap(uc.publicGate.SendMessageAsync(ctx, "global", actorMsg))
}
