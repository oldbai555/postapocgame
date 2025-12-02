package fuben

import (
	"context"

	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// SettleDungeonUseCase 副本结算用例
type SettleDungeonUseCase struct {
	playerRepo    repository.PlayerRepository
	levelUseCase  interfaces.LevelUseCase
	rewardUseCase interfaces.RewardUseCase
}

// NewSettleDungeonUseCase 创建副本结算用例
func NewSettleDungeonUseCase(
	playerRepo repository.PlayerRepository,
) *SettleDungeonUseCase {
	return &SettleDungeonUseCase{
		playerRepo: playerRepo,
	}
}

// SetDependencies 设置依赖（用于注入 LevelUseCase 和 RewardUseCase）
func (uc *SettleDungeonUseCase) SetDependencies(levelUseCase interfaces.LevelUseCase, rewardUseCase interfaces.RewardUseCase) {
	uc.levelUseCase = levelUseCase
	uc.rewardUseCase = rewardUseCase
}

// Execute 执行副本结算用例
// 注意：rewards 参数来自 DungeonServer 的结算结果，对应 jsonconf.DungeonReward 的精简结构
func (uc *SettleDungeonUseCase) Execute(ctx context.Context, roleID uint64, dungeonID uint32, difficulty uint32, success bool, rewards []*jsonconf.DungeonReward) error {
	// 如果副本失败，只更新记录，不发放奖励
	if !success {
		log.Infof("Dungeon failed, no rewards: RoleID=%d, DungeonID=%d", roleID, dungeonID)
		return nil
	}

	// 更新副本记录
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	if binaryData.DungeonData != nil {
		// 更新进入记录（结算时也算一次进入）
		// TODO: 这里可能需要区分结算和进入的逻辑
		_ = binaryData.DungeonData
	}

	// 转换奖励格式并发放
	if len(rewards) > 0 {
		rewardItems := make([]*jsonconf.ItemAmount, 0, len(rewards))
		for _, reward := range rewards {
			if reward == nil {
				continue
			}
			// 根据奖励类型转换
			switch reward.Type {
			case 1: // 经验奖励
				// 经验奖励通过等级系统发放
				if uc.levelUseCase != nil {
					if err := uc.levelUseCase.AddExp(ctx, roleID, uint64(reward.Count)); err != nil {
						log.Errorf("AddExp failed: %v", err)
						// 不返回错误，继续处理其他奖励
					}
				}
				continue // 经验已处理，跳过
			case 2: // 金币奖励
				rewardItems = append(rewardItems, &jsonconf.ItemAmount{
					ItemType: uint32(protocol.ItemType_ItemTypeMoney),
					ItemId:   reward.ItemID,
					Count:    int64(reward.Count),
					Bind:     1, // 副本奖励默认绑定
				})
			case 3: // 物品奖励
				rewardItems = append(rewardItems, &jsonconf.ItemAmount{
					ItemType: uint32(protocol.ItemType_ItemTypeMaterial),
					ItemId:   reward.ItemID,
					Count:    int64(reward.Count),
					Bind:     1, // 副本奖励默认绑定
				})
			default:
				log.Warnf("Unknown reward type: %d", reward.Type)
				continue
			}
		}

		// 发放奖励
		if len(rewardItems) > 0 && uc.rewardUseCase != nil {
			if err := uc.rewardUseCase.GrantRewards(ctx, roleID, rewardItems); err != nil {
				log.Errorf("GrantRewards failed: %v", err)
				return customerr.Wrap(err)
			}
		}
	}

	log.Infof("Dungeon settled successfully: RoleID=%d, DungeonID=%d", roleID, dungeonID)
	return nil
}
