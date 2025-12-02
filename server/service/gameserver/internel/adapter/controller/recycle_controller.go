package controller

import (
	"context"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/adapter/system"
	"postapocgame/server/service/gameserver/internel/di"
)

// RecycleController 回收系统控制器
type RecycleController struct {
	presenter *presenter.RecyclePresenter
}

// NewRecycleController 创建回收系统控制器
func NewRecycleController() *RecycleController {
	container := di.GetContainer()
	return &RecycleController{
		presenter: presenter.NewRecyclePresenter(container.NetworkGateway()),
	}
}

// HandleRecycleItem 处理回收物品请求
func (c *RecycleController) HandleRecycleItem(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SRecycleItemReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	if req.Count == 0 {
		req.Count = 1
	}

	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	recycleSys := system.GetRecycleSys(ctx)
	if recycleSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "recycle system not ready")
	}

	awards, err := recycleSys.RecycleItem(ctx, roleID, req.ItemId, req.Count)
	resp := &protocol.S2CRecycleItemResultReq{
		Success:       err == nil,
		ItemId:        req.ItemId,
		RecycledCount: req.Count,
		Awards:        awards,
	}
	if err != nil {
		resp.Message = err.Error()
	} else {
		resp.Message = "回收成功"
		presenter.PushBagData(ctx, sessionID)
		presenter.PushMoneyData(ctx, sessionID)
	}

	if sendErr := c.presenter.SendRecycleResult(ctx, sessionID, resp); sendErr != nil {
		return sendErr
	}
	return err
}
