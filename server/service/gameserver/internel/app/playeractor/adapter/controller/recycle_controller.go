package controller

import (
	"context"
	presenter2 "postapocgame/server/service/gameserver/internel/app/playeractor/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/system"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/gshare"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
)

// RecycleController 回收系统控制器
type RecycleController struct {
	presenter *presenter2.RecyclePresenter
}

// NewRecycleController 创建回收系统控制器
func NewRecycleController() *RecycleController {
	return &RecycleController{
		presenter: presenter2.NewRecyclePresenter(deps.NetworkGateway()),
	}
}

// HandleRecycleItem 处理回收物品请求
func (c *RecycleController) HandleRecycleItem(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := gshare.GetSessionIDFromContext(ctx)
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

	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 检查系统是否开启
	recycleSys := system.GetRecycleSys(ctx)
	if recycleSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "回收系统未开启")
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
		presenter2.PushBagData(ctx, sessionID)
		presenter2.PushMoneyData(ctx, sessionID)
	}

	if sendErr := c.presenter.SendRecycleResult(ctx, sessionID, resp); sendErr != nil {
		return sendErr
	}
	return err
}
