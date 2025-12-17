package equip

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/bag"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// EquipController 装备控制器
type EquipController struct {
	deps             Deps
	equipItemUseCase *EquipItemUseCase
	unEquipUseCase   *UnEquipItemUseCase
	presenter        *EquipPresenter
}

// NewEquipController 创建装备控制器
func NewEquipController(rt *runtime.Runtime) *EquipController {
	d := depsFromRuntime(rt)
	return &EquipController{
		deps:             d,
		equipItemUseCase: NewEquipItemUseCase(d),
		unEquipUseCase:   NewUnEquipItemUseCase(d),
		presenter:        NewEquipPresenter(d.NetworkGateway),
	}
}

// HandleEquipItem 处理装备物品请求
func (c *EquipController) HandleEquipItem(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	equipSys := GetEquipSys(ctx)
	if equipSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "装备系统未开启")
	}

	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SEquipItemReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 执行装备物品用例
	err = c.equipItemUseCase.Execute(ctx, roleID, req.ItemId, req.Slot)

	// 构建响应
	resp := &protocol.S2CEquipResultReq{
		Slot:   req.Slot,
		ItemId: req.ItemId,
	}

	// 发送响应
	if sendErr := c.presenter.SendEquipResult(ctx, sessionID, resp); sendErr != nil {
		return sendErr
	}

	// 如果成功，推送背包数据更新
	if err != nil {
		return customerr.Wrap(err)
	}

	// 通过 Bag 系统推送当前背包数据
	bagSys := bag.GetBagSys(ctx)
	if bagSys == nil {
		log.Warnf("get bag sys failed")
		return customerr.Wrap(err)
	}
	bagData, bagErr := bagSys.GetBagData(ctx)
	if bagErr != nil {
		log.Warnf("get bag data failed: %v", bagErr)
		return customerr.Wrap(err)
	}

	if err := bag.NewBagPresenter(c.deps.NetworkGateway).SendBagData(ctx, sessionID, bagData); err != nil {
		log.Warnf("send bag data failed: %v", err)
	}

	return nil
}

// init 注册协议与事件
// init() 函数已移除，注册逻辑迁移至 playeractor/register.RegisterAll()
