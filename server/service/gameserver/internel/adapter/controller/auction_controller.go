package controller

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/di"
	auctionusecase "postapocgame/server/service/gameserver/internel/usecase/auction"
)

// AuctionController 拍卖行控制器
type AuctionController struct {
	putOnUC   *auctionusecase.PutOnUseCase
	buyUC     *auctionusecase.BuyUseCase
	queryUC   *auctionusecase.QueryUseCase
	presenter *presenter.AuctionPresenter
}

func NewAuctionController() *AuctionController {
	container := di.GetContainer()
	publicActor := container.PublicActorGateway()
	return &AuctionController{
		putOnUC:   auctionusecase.NewPutOnUseCase(publicActor),
		buyUC:     auctionusecase.NewBuyUseCase(publicActor),
		queryUC:   auctionusecase.NewQueryUseCase(publicActor),
		presenter: presenter.NewAuctionPresenter(container.NetworkGateway()),
	}
}

func (c *AuctionController) HandlePutOn(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	var req protocol.C2SAuctionPutOnReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}
	if err := c.putOnUC.Execute(ctx, roleID, &req); err != nil {
		return c.presenter.SendError(ctx, sessionID, err.Error())
	}
	return nil
}

func (c *AuctionController) HandleBuy(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	var req protocol.C2SAuctionBuyReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}
	if err := c.buyUC.Execute(ctx, roleID, req.AuctionId); err != nil {
		return c.presenter.SendError(ctx, sessionID, err.Error())
	}
	return nil
}

func (c *AuctionController) HandleQuery(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}
	var req protocol.C2SAuctionQueryReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}
	if err := c.queryUC.Execute(ctx, sessionID, &req); err != nil {
		return c.presenter.SendError(ctx, sessionID, err.Error())
	}
	return nil
}
