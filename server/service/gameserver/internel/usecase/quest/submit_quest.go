package quest

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// SubmitQuestUseCase 提交任务用例
type SubmitQuestUseCase struct {
	playerRepo           repository.PlayerRepository
	configManager        interfaces.ConfigManager
	levelUseCase         interfaces.LevelUseCase
	rewardUseCase        interfaces.RewardUseCase
	dailyActivityUseCase interfaces.DailyActivityUseCase
}

// NewSubmitQuestUseCase 创建提交任务用例
func NewSubmitQuestUseCase(
	playerRepo repository.PlayerRepository,
	configManager interfaces.ConfigManager,
) *SubmitQuestUseCase {
	return &SubmitQuestUseCase{
		playerRepo:    playerRepo,
		configManager: configManager,
	}
}

// SetDependencies 设置依赖（用于注入 LevelUseCase、RewardUseCase、DailyActivityUseCase）
func (uc *SubmitQuestUseCase) SetDependencies(
	levelUseCase interfaces.LevelUseCase,
	rewardUseCase interfaces.RewardUseCase,
	dailyActivityUseCase interfaces.DailyActivityUseCase,
) {
	uc.levelUseCase = levelUseCase
	uc.rewardUseCase = rewardUseCase
	uc.dailyActivityUseCase = dailyActivityUseCase
}

// Execute 执行提交任务用例
func (uc *SubmitQuestUseCase) Execute(ctx context.Context, roleID uint64, questId uint32) error {
	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	// 获取任务
	quest, questType := uc.getQuestWithType(binaryData.QuestData, questId)
	if quest == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "quest not found: %d", questId)
	}

	// 检查任务是否完成
	if !uc.isQuestCompleted(binaryData.QuestData, questId) {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "quest not completed: %d", questId)
	}

	// 获取任务配置
	questConfigRaw, ok := uc.configManager.GetQuestConfig(questId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "quest config not found: %d", questId)
	}

	questConfig, ok := questConfigRaw.(*jsonconf.QuestConfig)
	if !ok || questConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "invalid quest config type")
	}

	// 发放经验奖励
	if questConfig.ExpReward > 0 && uc.levelUseCase != nil {
		if err := uc.levelUseCase.AddExp(ctx, roleID, uint64(questConfig.ExpReward)); err != nil {
			log.Errorf("AddExp failed: %v", err)
			// 经验发放失败不影响任务提交，只记录日志
		}
	}

	// 发放物品奖励
	if len(questConfig.Rewards) > 0 && uc.rewardUseCase != nil {
		rewards := make([]*jsonconf.ItemAmount, 0, len(questConfig.Rewards))
		for _, reward := range questConfig.Rewards {
			rewards = append(rewards, &jsonconf.ItemAmount{
				ItemType: uint32(reward.Type),
				ItemId:   reward.ItemId,
				Count:    int64(reward.Count),
				Bind:     1, // 任务奖励默认绑定
			})
		}
		if err := uc.rewardUseCase.GrantRewards(ctx, roleID, rewards); err != nil {
			log.Errorf("GrantRewards failed: %v", err)
			return customerr.Wrap(err)
		}
	}

	// 日常任务奖励活跃点
	if questType == questCategoryDaily && questConfig.ActivePoint > 0 && uc.dailyActivityUseCase != nil {
		if err := uc.dailyActivityUseCase.AddActivePoints(ctx, roleID, questConfig.ActivePoint); err != nil {
			log.Warnf("AddActivePoints failed: RoleID=%d, QuestID=%d, Err=%v", roleID, questId, err)
		}
	}

	// 处理任务完成后的逻辑
	switch questType {
	case questCategoryDaily, questCategoryWeekly:
		// 日常/周常任务：支持在同一自然日/周内多次完成，受 MaxCount 限制
		if binaryData.QuestData.QuestFinishCount == nil {
			binaryData.QuestData.QuestFinishCount = make(map[uint32]uint32)
		}
		finishCount := binaryData.QuestData.QuestFinishCount[questId] + 1
		binaryData.QuestData.QuestFinishCount[questId] = finishCount

		if questConfig.MaxCount > 0 && finishCount >= questConfig.MaxCount {
			// 达到最大完成次数后，从当前桶中移除，等待下一次刷新重新生成
			uc.removeQuest(binaryData.QuestData, questId)
		} else {
			// 未达到最大次数，重置进度，允许同一自然日/周内再次完成
			uc.resetQuestProgress(quest)
		}
	default:
		uc.removeQuest(binaryData.QuestData, questId)
	}

	// 检查是否有后续任务（任务链）
	if len(questConfig.NextQuests) > 0 {
		acceptQuestUC := NewAcceptQuestUseCase(uc.playerRepo, uc.configManager)
		acceptQuestUC.SetDependencies(uc.levelUseCase)
		for _, nextQuestId := range questConfig.NextQuests {
			// 自动接取后续任务
			if err := acceptQuestUC.Execute(ctx, roleID, nextQuestId); err != nil {
				log.Warnf("Auto accept next quest failed: QuestID=%d, NextQuestID=%d, Error=%v", questId, nextQuestId, err)
			} else {
				log.Infof("Auto accepted next quest: QuestID=%d, NextQuestID=%d", questId, nextQuestId)
			}
		}
	}

	log.Infof("Quest submitted: RoleID=%d, QuestID=%d", roleID, questId)
	return nil
}

// getQuestWithType 获取任务及其类型
func (uc *SubmitQuestUseCase) getQuestWithType(questData *protocol.SiQuestData, questId uint32) (*protocol.QuestData, uint32) {
	if questData == nil || questData.QuestMap == nil {
		return nil, 0
	}
	for questType, bucket := range questData.QuestMap {
		if bucket == nil {
			continue
		}
		for _, quest := range bucket.Quests {
			if quest != nil && quest.Id == questId {
				return quest, questType
			}
		}
	}
	return nil, 0
}

// isQuestCompleted 检查任务是否完成
func (uc *SubmitQuestUseCase) isQuestCompleted(questData *protocol.SiQuestData, questId uint32) bool {
	quest := uc.getQuest(questData, questId)
	if quest == nil {
		return false
	}

	questConfigRaw, ok := uc.configManager.GetQuestConfig(questId)
	if !ok {
		return false
	}

	questConfig, ok := questConfigRaw.(*jsonconf.QuestConfig)
	if !ok || questConfig == nil {
		return false
	}

	// 检查所有目标是否完成
	for i, target := range questConfig.Targets {
		if i >= len(quest.Progress) {
			return false
		}
		if quest.Progress[i] < target.Count {
			return false
		}
	}

	return true
}

// getQuest 获取指定任务
func (uc *SubmitQuestUseCase) getQuest(questData *protocol.SiQuestData, questId uint32) *protocol.QuestData {
	if questData == nil || questData.QuestMap == nil {
		return nil
	}
	for _, bucket := range questData.QuestMap {
		if bucket == nil {
			continue
		}
		for _, quest := range bucket.Quests {
			if quest != nil && quest.Id == questId {
				return quest
			}
		}
	}
	return nil
}

// removeQuest 移除任务
func (uc *SubmitQuestUseCase) removeQuest(questData *protocol.SiQuestData, questId uint32) {
	if questData == nil || questData.QuestMap == nil {
		return
	}

	for questType, bucket := range questData.QuestMap {
		if bucket == nil || len(bucket.Quests) == 0 {
			continue
		}
		for i, quest := range bucket.Quests {
			if quest != nil && quest.Id == questId {
				bucket.Quests = append(bucket.Quests[:i], bucket.Quests[i+1:]...)
				if len(bucket.Quests) == 0 {
					bucket.Quests = make([]*protocol.QuestData, 0)
				}
				log.Debugf("Quest removed: QuestID=%d, Type=%d", questId, questType)
				return
			}
		}
	}
}

// resetQuestProgress 重置任务进度
func (uc *SubmitQuestUseCase) resetQuestProgress(quest *protocol.QuestData) {
	if quest == nil {
		return
	}
	for i := range quest.Progress {
		quest.Progress[i] = 0
	}
}
