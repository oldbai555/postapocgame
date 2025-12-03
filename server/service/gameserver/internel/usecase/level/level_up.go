package level

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// LevelUpUseCase 等级提升用例
type LevelUpUseCase struct {
	playerRepo     repository.PlayerRepository
	eventPublisher interfaces.EventPublisher
	configManager  interfaces.ConfigManager
	// 依赖其他系统（暂时通过接口定义，后续重构时注入）
	moneyUseCase interfaces.MoneyUseCase // 可选
	attrUseCase  interfaces.AttrUseCase  // 可选
}

// NewLevelUpUseCase 创建等级提升用例
func NewLevelUpUseCase(
	playerRepo repository.PlayerRepository,
	eventPublisher interfaces.EventPublisher,
	configManager interfaces.ConfigManager,
) *LevelUpUseCase {
	return &LevelUpUseCase{
		playerRepo:     playerRepo,
		eventPublisher: eventPublisher,
		configManager:  configManager,
	}
}

// SetDependencies 设置依赖（可选，用于后续系统重构后注入）
func (uc *LevelUpUseCase) SetDependencies(moneyUseCase interfaces.MoneyUseCase, attrUseCase interfaces.AttrUseCase) {
	uc.moneyUseCase = moneyUseCase
	uc.attrUseCase = attrUseCase
}

// Execute 执行等级提升检查（循环检查，可能连续升级）
func (uc *LevelUpUseCase) Execute(ctx context.Context, roleID uint64) error {
	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	if binaryData.LevelData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "level data not initialized")
	}

	// 循环检查升级（可能一次获得大量经验，连续升级）
	for {
		// 获取当前等级的配置
		levelConfigRaw, ok := uc.configManager.GetLevelConfig(binaryData.LevelData.Level)
		if !ok {
			// 没有更高等级的配置，已达到最高等级
			break
		}

		levelConfig, ok := levelConfigRaw.(*jsonconf.LevelConfig)
		if !ok {
			log.Errorf("invalid level config type: level=%d", binaryData.LevelData.Level)
			break
		}

		// 检查是否满足升级条件
		if binaryData.LevelData.Exp < int64(levelConfig.ExpNeeded) {
			break
		}

		// 扣除升级所需经验
		binaryData.LevelData.Exp -= int64(levelConfig.ExpNeeded)

		// 同步更新货币系统中的经验值
		if uc.moneyUseCase != nil {
			// 新方式：通过接口调用
			if err := uc.moneyUseCase.UpdateExp(ctx, roleID, binaryData.LevelData.Exp); err != nil {
				log.Errorf("update exp in money system failed: %v", err)
			}
		} else {
			// 旧方式：直接更新 BinaryData（向后兼容）
			if binaryData.MoneyData != nil {
				if binaryData.MoneyData.MoneyMap == nil {
					binaryData.MoneyData.MoneyMap = make(map[uint32]int64)
				}
				expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
				binaryData.MoneyData.MoneyMap[expMoneyID] = binaryData.LevelData.Exp
			}
		}

		// 升级
		binaryData.LevelData.Level++

		playerRole := adaptercontext.MustGetPlayerRoleFromContext(ctx)
		if playerRole != nil {
			log.Infof("Player level up: PlayerID=%d, NewLevel=%d, RemainingExp=%d",
				roleID, binaryData.LevelData.Level, binaryData.LevelData.Exp)
		}

		// 发放升级奖励
		if len(levelConfig.Rewards) > 0 {
			rewards := make([]*jsonconf.ItemAmount, 0, len(levelConfig.Rewards))
			for _, reward := range levelConfig.Rewards {
				rewards = append(rewards, &jsonconf.ItemAmount{
					ItemType: uint32(reward.Type),
					ItemId:   reward.ItemId,
					Count:    int64(reward.Count),
					Bind:     1, // 升级奖励默认绑定
				})
			}
			// 通过 PlayerRole 发放奖励（暂时保持旧方式，等后续重构）
			// 注意：这里通过 Context 获取 PlayerRole，违反了 Clean Architecture 原则
			// 等 RewardUseCase 接口实现后，应该通过接口调用
			playerRole := adaptercontext.MustGetPlayerRoleFromContext(ctx)
			if playerRole != nil {
				if err := playerRole.GrantRewards(ctx, rewards); err != nil {
					log.Errorf("Grant level up rewards failed: %v", err)
					// 奖励发放失败不影响升级，只记录日志
				}
			}
		}

		// 发布升级事件
		uc.eventPublisher.PublishPlayerEvent(ctx, gevent.OnPlayerLevelUp, map[string]interface{}{
			"level": binaryData.LevelData.Level,
		})
	}

	// 标记属性系统需要重算
	if uc.attrUseCase != nil {
		// 通过接口调用（Clean Architecture 方式）
		if err := uc.attrUseCase.MarkDirty(ctx, roleID, uint32(protocol.SaAttrSys_SaLevel)); err != nil {
			log.Errorf("mark attr dirty failed: %v", err)
		}
	} else {
		// 向后兼容：通过 Context 获取 AttrCalculator（如果 AttrUseCase 未注入）
		// 注意：这种方式违反了 Clean Architecture 原则，应该通过接口调用
		// 但在过渡期，为了保持功能正常，保留此逻辑
		playerRole := adaptercontext.MustGetPlayerRoleFromContext(ctx)
		if playerRole != nil {
			// 通过 PlayerRole 获取 AttrCalculator
			// 注意：这里需要导入 entity 包，但为了避免循环依赖，暂时不处理
			// 等所有系统重构完成后，统一移除此逻辑
			_ = playerRole
		}
	}

	return nil
}
