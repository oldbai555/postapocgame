package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/equip"
)

// EquipSystemAdapter 装备系统适配器
//
// 生命周期职责：
// - OnInit: 调用 InitEquipDataUseCase 初始化装备数据（装备列表结构）
// - 其他生命周期: 暂未使用
//
// 业务逻辑：所有业务逻辑（装备/卸下、属性加成）均在 UseCase 层实现
// 事件订阅：在 equip_system_adapter_init.go 中订阅 OnEquipChange/OnEquipUpgrade 事件，标记属性系统需要重算
//
// ⚠️ 防退化机制：禁止在 SystemAdapter 中编写业务规则逻辑，只允许调用 UseCase 与管理生命周期
type EquipSystemAdapter struct {
	*BaseSystemAdapter
	equipItemUseCase     *equip.EquipItemUseCase
	unEquipItemUseCase   *equip.UnEquipItemUseCase
	initEquipDataUseCase *equip.InitEquipDataUseCase
}

// NewEquipSystemAdapter 创建装备系统适配器
func NewEquipSystemAdapter() *EquipSystemAdapter {
	container := di.GetContainer()
	equipItemUC := equip.NewEquipItemUseCase(container.PlayerGateway(), container.EventPublisher(), container.ConfigGateway())
	unEquipItemUC := equip.NewUnEquipItemUseCase(container.PlayerGateway(), container.EventPublisher())

	return &EquipSystemAdapter{
		BaseSystemAdapter:    NewBaseSystemAdapter(uint32(protocol.SystemId_SysEquip)),
		equipItemUseCase:     equipItemUC,
		unEquipItemUseCase:   unEquipItemUC,
		initEquipDataUseCase: equip.NewInitEquipDataUseCase(container.PlayerGateway()),
	}
}

// OnInit 系统初始化
func (a *EquipSystemAdapter) OnInit(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("equip sys OnInit get role err:%v", err)
		return
	}
	// 初始化装备数据（包括装备列表结构等业务逻辑）
	if err := a.initEquipDataUseCase.Execute(ctx, roleID); err != nil {
		log.Errorf("equip sys OnInit init equip data err:%v", err)
		return
	}
	// 获取装备数量用于日志（可选）
	binaryData, _ := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	equipCount := 0
	if binaryData != nil && binaryData.EquipData != nil {
		equipCount = len(binaryData.EquipData.Equips)
	}
	log.Infof("EquipSys initialized: PlayerID=%d, EquipCount=%d", roleID, equipCount)
}

// GetEquipData 获取装备数据（用于协议）
func (a *EquipSystemAdapter) GetEquipData(ctx context.Context) *protocol.SiEquipData {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil || binaryData == nil {
		return nil
	}
	return binaryData.EquipData
}
