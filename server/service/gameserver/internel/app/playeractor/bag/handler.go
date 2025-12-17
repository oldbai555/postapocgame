package bag

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// BagController 背包控制器
type BagController struct {
	addItemUseCase    *AddItemUseCase
	removeItemUseCase *RemoveItemUseCase
	hasItemUseCase    *HasItemUseCase
	presenter         *BagPresenter
}

// NewBagController 创建背包控制器
func NewBagController(rt *runtime.Runtime) *BagController {
	d := depsFromRuntime(rt)
	return &BagController{
		addItemUseCase:    NewAddItemUseCase(d),
		removeItemUseCase: NewRemoveItemUseCase(d),
		hasItemUseCase:    NewHasItemUseCase(d),
		presenter:         NewBagPresenter(d.NetworkGateway),
	}
}

// HandleOpenBag 处理打开背包请求
func (c *BagController) HandleOpenBag(ctx context.Context, _ *network.ClientMessage) error {
	// 检查系统是否开启
	bagSys := GetBagSys(ctx)
	if bagSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "背包系统未开启")
	}

	sessionId, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 从 System 获取背包数据
	bagData, err := bagSys.GetBagData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 通过 Presenter 发送响应
	return c.presenter.SendBagData(ctx, sessionId, bagData)
}

// HandleAddItem 处理添加物品请求（RPC，来自 DungeonServer）
func (c *BagController) HandleAddItem(ctx context.Context, sessionID string, data []byte) error {
	// 检查系统是否开启
	bagSys := GetBagSys(ctx)
	if bagSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "背包系统未开启")
	}

	var req protocol.D2GAddItemReq
	if err := proto.Unmarshal(data, &req); err != nil {
		return customerr.Wrap(err)
	}

	// 执行添加物品用例
	err := c.addItemUseCase.Execute(ctx, req.RoleId, req.ItemId, req.Count, 0) // bind=0 表示非绑定
	if err != nil {
		return customerr.Wrap(err)
	}

	// 从 System 获取更新后的背包数据并发送
	bagData, err := bagSys.GetBagData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 通过 Presenter 发送背包数据更新
	return c.presenter.SendBagData(ctx, sessionID, bagData)
}

// init() 函数已移除，注册逻辑迁移至 playeractor/register.RegisterAll()
