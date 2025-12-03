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
	"postapocgame/server/service/gameserver/internel/usecase/equip"
)

// EquipController 装备控制器
type EquipController struct {
	equipItemUseCase   *equip.EquipItemUseCase
	unEquipItemUseCase *equip.UnEquipItemUseCase
	presenter          *presenter.EquipPresenter
}

// NewEquipController 创建装备控制器
func NewEquipController() *EquipController {
	container := di.GetContainer()
	equipItemUC := equip.NewEquipItemUseCase(container.PlayerGateway(), container.EventPublisher(), container.ConfigGateway())
	unEquipItemUC := equip.NewUnEquipItemUseCase(container.PlayerGateway(), container.EventPublisher())

	// 注入 BagUseCase 依赖（通过适配器）
	bagUseCase := system.NewBagUseCaseAdapter()
	equipItemUC.SetDependencies(bagUseCase, nil) // rewardUseCase 暂时为 nil
	unEquipItemUC.SetDependencies(bagUseCase)

	return &EquipController{
		equipItemUseCase:   equipItemUC,
		unEquipItemUseCase: unEquipItemUC,
		presenter:          presenter.NewEquipPresenter(container.NetworkGateway()),
	}
}

// HandleEquipItem 处理装备物品请求
func (c *EquipController) HandleEquipItem(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	equipSys := system.GetEquipSys(ctx)
	if equipSys == nil {
		sessionID, _ := adaptercontext.GetSessionIDFromContext(ctx)
		resp := &protocol.S2CEquipResultReq{
			Slot:    0,
			ItemId:  0,
			ErrCode: uint32(protocol.ErrorCode_System_NotEnabled),
		}
		if sendErr := c.presenter.SendEquipResult(ctx, sessionID, resp); sendErr != nil {
			return sendErr
		}
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "装备系统未开启")
	}

	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SEquipItemReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 执行装备物品用例
	err = c.equipItemUseCase.Execute(ctx, roleID, req.ItemId, req.Slot)

	// 构建响应
	resp := &protocol.S2CEquipResultReq{
		Slot:    req.Slot,
		ItemId:  req.ItemId,
		ErrCode: errCodeFromError(err),
	}

	// 发送响应
	if sendErr := c.presenter.SendEquipResult(ctx, sessionID, resp); sendErr != nil {
		return sendErr
	}

	// 如果成功，推送背包数据更新
	if err == nil {
		// 通过 BagPresenter 推送背包数据
		binaryData, _ := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
		if binaryData != nil && binaryData.BagData != nil {
			bagPresenter := presenter.NewBagPresenter(di.GetContainer().NetworkGateway())
			_ = bagPresenter.SendBagData(ctx, sessionID, binaryData.BagData)
		}
	}

	return err
}

// errCodeFromError 从错误中提取错误码
func errCodeFromError(err error) uint32 {
	if err == nil {
		return uint32(protocol.ErrorCode_Success)
	}
	// 这里需要根据错误类型提取错误码
	// 暂时返回通用错误
	return uint32(protocol.ErrorCode_Internal_Error)
}
