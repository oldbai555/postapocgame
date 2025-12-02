package auction

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// QueryUseCase 查询拍卖
type QueryUseCase struct {
	publicGate interfaces.PublicActorGateway
}

func NewQueryUseCase(publicGate interfaces.PublicActorGateway) *QueryUseCase {
	return &QueryUseCase{publicGate: publicGate}
}

func (uc *QueryUseCase) Execute(ctx context.Context, sessionID string, req *protocol.C2SAuctionQueryReq) error {
	if sessionID == "" {
		return customerr.NewError("session 无效")
	}
	query := &protocol.AuctionQueryMsg{
		ItemId:             req.GetItemId(),
		Page:               req.GetPage(),
		PageSize:           req.GetPageSize(),
		RequesterSessionId: sessionID,
	}
	data, err := proto.Marshal(query)
	if err != nil {
		return customerr.Wrap(err)
	}
	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAuctionQuery), data)
	return customerr.Wrap(uc.publicGate.SendMessageAsync(ctx, "global", actorMsg))
}
