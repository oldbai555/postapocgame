package recycle

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/presenter"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// RecycleController 回收系统控制器
type RecycleController struct {
	presenter *RecyclePresenter
}

// NewRecycleController 创建回收系统控制器
func NewRecycleController(rt *runtime.Runtime) *RecycleController {
	d := depsFromRuntime(rt)
	return &RecycleController{
		presenter: NewRecyclePresenter(d.NetworkGateway),
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
	recycleSys := GetRecycleSys(ctx)
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
		presenter.PushBagData(ctx, sessionID)
		presenter.PushMoneyData(ctx, sessionID)
	}

	if sendErr := c.presenter.SendRecycleResult(ctx, sessionID, resp); sendErr != nil {
		return sendErr
	}
	return err
}

// init() 函数已移除，注册逻辑迁移至 playeractor/register.RegisterAll()
