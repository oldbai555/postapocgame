package item_use

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

const (
	// DefaultItemUseCooldownSeconds 默认物品使用冷却时间（秒）
	DefaultItemUseCooldownSeconds int64 = 5
)

// UseItemUseCase 使用物品用例
type UseItemUseCase struct {
	playerRepo     repository.PlayerRepository
	configManager  interfaces.ConfigManager
	bagUseCase     interfaces.BagUseCase
	levelUseCase   interfaces.LevelUseCase // 用于添加经验
	dungeonGateway interfaces.DungeonServerGateway
}

// NewUseItemUseCase 创建使用物品用例
func NewUseItemUseCase(
	playerRepo repository.PlayerRepository,
	configManager interfaces.ConfigManager,
	dungeonGateway interfaces.DungeonServerGateway,
) *UseItemUseCase {
	return &UseItemUseCase{
		playerRepo:     playerRepo,
		configManager:  configManager,
		dungeonGateway: dungeonGateway,
	}
}

// SetDependencies 设置依赖（用于注入 BagUseCase 和 LevelUseCase）
func (uc *UseItemUseCase) SetDependencies(bagUseCase interfaces.BagUseCase, levelUseCase interfaces.LevelUseCase) {
	uc.bagUseCase = bagUseCase
	uc.levelUseCase = levelUseCase
}

// Execute 执行使用物品用例
func (uc *UseItemUseCase) Execute(ctx context.Context, roleID uint64, itemID uint32, count uint32) error {
	// 默认使用数量为1
	if count == 0 {
		count = 1
	}

	// 检查物品配置
	itemConfigRaw, ok := uc.configManager.GetItemConfig(itemID)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config not found: %d", itemID)
	}

	itemConfig, ok := itemConfigRaw.(*jsonconf.ItemConfig)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "invalid item config type")
	}

	// 检查物品是否可使用（通过Flag检查）
	if itemConfig.Flag&uint64(protocol.ItemFlag_ItemFlagCanUse) == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item cannot be used")
	}

	// 检查物品类型（只有消耗品可以使用）
	if itemConfig.Type != uint32(protocol.ItemType_ItemTypeConsume) {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "only consume items can be used")
	}

	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	// 检查冷却时间
	now := servertime.Now().Unix()
	if binaryData.ItemUseData == nil || binaryData.ItemUseData.CooldownMap == nil {
		binaryData.ItemUseData = &protocol.SiItemUseData{
			CooldownMap: make(map[uint32]int64),
		}
	}
	if cooldownEnd, exists := binaryData.ItemUseData.CooldownMap[itemID]; exists && cooldownEnd > now {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item is in cooldown")
	}

	// 检查背包中是否有该物品
	if uc.bagUseCase != nil {
		hasItem, err := uc.bagUseCase.HasItem(ctx, roleID, itemID, count)
		if err != nil {
			return err
		}
		if !hasItem {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not enough")
		}
	} else {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag use case not initialized")
	}

	// 获取物品使用效果配置
	useEffectConfigRaw, ok := uc.configManager.GetItemUseEffectConfig(itemID)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item use effect config not found: %d", itemID)
	}

	useEffectConfig, ok := useEffectConfigRaw.(*jsonconf.ItemUseEffectConfig)
	if !ok || useEffectConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "invalid item use effect config type")
	}

	// 应用物品效果
	var hpDelta int64 = 0
	var mpDelta int64 = 0
	var expDelta int64 = 0

	for i := uint32(0); i < count; i++ {
		// 遍历效果值数组，应用每个效果
		for _, value := range useEffectConfig.Values {
			switch useEffectConfig.EffectType {
			case 1: // 恢复HP
				hpDelta += int64(value)
			case 2: // 恢复MP
				mpDelta += int64(value)
			case 3: // 增加经验
				expDelta += int64(value)
			}
		}
	}

	// 当前设计下，战斗中的 HP/MP 以 DungeonServer 为权威，GameServer 不主动回写战斗内血蓝。
	// 物品产生的 HP/MP 效果主要用于副本外（例如恢复药水在非战斗场景生效），由 DungeonServer 协议自行处理；
	// 因此此处仅记录数值变化用于日志和后续扩展，不直接向 DungeonServer 发送同步请求。
	_ = hpDelta
	_ = mpDelta

	// 如果有经验变化，通过 LevelUseCase 添加经验
	if expDelta > 0 && uc.levelUseCase != nil {
		if err := uc.levelUseCase.AddExp(ctx, roleID, uint64(expDelta)); err != nil {
			log.Errorf("add exp failed: %v", err)
			// 不返回错误，继续执行
		}
	}

	// 扣除物品数量
	if err := uc.bagUseCase.RemoveItem(ctx, roleID, itemID, count); err != nil {
		return err
	}

	// 设置冷却时间（默认5秒，可以根据配置调整）
	if binaryData.ItemUseData.CooldownMap == nil {
		binaryData.ItemUseData.CooldownMap = make(map[uint32]int64)
	}
	binaryData.ItemUseData.CooldownMap[itemID] = now + DefaultItemUseCooldownSeconds

	log.Infof("Item used: RoleID=%d, ItemID=%d, Count=%d, HPDelta=%d, MPDelta=%d, ExpDelta=%d", roleID, itemID, count, hpDelta, mpDelta, expDelta)

	return nil
}
