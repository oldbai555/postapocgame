package level

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

// LevelUpUseCase 等级提升用例
type LevelUpUseCase struct {
	playerRepo     iface.PlayerRepository
	eventPublisher iface.EventPublisher
	configManager  iface.ConfigManager
	// 依赖其他系统（通过接口注入）
	moneyUseCase  iface.MoneyUseCase  // 必须：同步经验货币
	rewardUseCase iface.RewardUseCase // 必须：升级奖励
}

// NewLevelUpUseCase 创建等级提升用例
func NewLevelUpUseCase(
	playerRepo iface.PlayerRepository,
	eventPublisher iface.EventPublisher,
	configManager iface.ConfigManager,
) *LevelUpUseCase {
	return &LevelUpUseCase{
		playerRepo:     playerRepo,
		eventPublisher: eventPublisher,
		configManager:  configManager,
	}
}

// SetDependencies 设置依赖
func (uc *LevelUpUseCase) SetDependencies(
	moneyUseCase iface.MoneyUseCase,
	rewardUseCase iface.RewardUseCase,
) {
	uc.moneyUseCase = moneyUseCase
	uc.rewardUseCase = rewardUseCase
}

// Execute 执行等级提升检查（循环检查，可能连续升级）
func (uc *LevelUpUseCase) Execute(ctx context.Context, roleID uint64) error {
	// 获取 BinaryData（共享引用）
	levelData, err := uc.playerRepo.GetLevelData(ctx)
	if err != nil {
		return err
	}

	// 循环检查升级（可能一次获得大量经验，连续升级）
	for {
		// 获取当前等级的配置
		levelConfig := uc.configManager.GetLevelConfig(levelData.Level)
		if levelConfig == nil {
			// 没有更高等级的配置，已达到最高等级
			break
		}

		// 检查是否满足升级条件
		if levelData.Exp < int64(levelConfig.ExpNeeded) {
			break
		}

		// 扣除升级所需经验
		levelData.Exp -= int64(levelConfig.ExpNeeded)

		// 同步更新货币系统中的经验值
		if uc.moneyUseCase == nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "money use case not injected")
		}
		if err := uc.moneyUseCase.UpdateExp(ctx, roleID, levelData.Exp); err != nil {
			return err
		}

		// 升级
		levelData.Level++

		playerRole := gshare.MustGetPlayerRoleFromContext(ctx)
		if playerRole != nil {
			log.Infof("Player level up: PlayerID=%d, NewLevel=%d, RemainingExp=%d", roleID, levelData.Level, levelData.Exp)
		}

		// 发放升级奖励
		if len(levelConfig.Rewards) > 0 {
			if uc.rewardUseCase == nil {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "reward use case not injected")
			}
			rewards := make([]*jsonconf.ItemAmount, 0, len(levelConfig.Rewards))
			for _, reward := range levelConfig.Rewards {
				rewards = append(rewards, &jsonconf.ItemAmount{
					ItemType: uint32(reward.Type),
					ItemId:   reward.ItemId,
					Count:    int64(reward.Count),
					Bind:     1, // 升级奖励默认绑定
				})
			}
			if err := uc.rewardUseCase.GrantRewards(ctx, roleID, rewards); err != nil {
				return err
			}
		}

		// 发布升级事件
		uc.eventPublisher.PublishPlayerEvent(ctx, gevent.OnPlayerLevelUp, map[string]interface{}{
			"level": levelData.Level,
		})
	}

	return nil
}
