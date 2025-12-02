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
type EquipSystemAdapter struct {
	*BaseSystemAdapter
	equipItemUseCase   *equip.EquipItemUseCase
	unEquipItemUseCase *equip.UnEquipItemUseCase
}

// NewEquipSystemAdapter 创建装备系统适配器
func NewEquipSystemAdapter() *EquipSystemAdapter {
	container := di.GetContainer()
	equipItemUC := equip.NewEquipItemUseCase(container.PlayerGateway(), container.EventPublisher(), container.ConfigGateway())
	unEquipItemUC := equip.NewUnEquipItemUseCase(container.PlayerGateway(), container.EventPublisher())

	return &EquipSystemAdapter{
		BaseSystemAdapter:  NewBaseSystemAdapter(uint32(protocol.SystemId_SysEquip)),
		equipItemUseCase:   equipItemUC,
		unEquipItemUseCase: unEquipItemUC,
	}
}

// OnInit 系统初始化
func (a *EquipSystemAdapter) OnInit(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("equip sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		log.Errorf("equip sys OnInit get binary data err:%v", err)
		return
	}

	// 如果equip_data不存在，则初始化
	if binaryData.EquipData == nil {
		binaryData.EquipData = &protocol.SiEquipData{
			Equips: make([]*protocol.EquipSt, 0),
		}
	}

	log.Infof("EquipSys initialized: PlayerID=%d, EquipCount=%d", roleID, len(binaryData.EquipData.Equips))
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
