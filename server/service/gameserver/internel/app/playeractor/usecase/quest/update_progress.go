package quest

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
)

// UpdateQuestProgressUseCase 更新任务进度用例
type UpdateQuestProgressUseCase struct {
	playerRepo    repository.PlayerRepository
	configManager interfaces.ConfigManager
}

// NewUpdateQuestProgressUseCase 创建更新任务进度用例
func NewUpdateQuestProgressUseCase(
	playerRepo repository.PlayerRepository,
	configManager interfaces.ConfigManager,
) *UpdateQuestProgressUseCase {
	return &UpdateQuestProgressUseCase{
		playerRepo:    playerRepo,
		configManager: configManager,
	}
}

// Execute 执行更新任务进度用例
// questType: 任务类型（1=和NPC对话，2=学习技能，3=击杀怪物）
// targetId: 目标ID（对于type=1，传入npcId；对于type=2和3，可以传入0表示任意）
// count: 增加的数量（默认1）
func (uc *UpdateQuestProgressUseCase) Execute(ctx context.Context, roleID uint64, questType uint32, targetId uint32, count uint32) error {
	questData, err := uc.playerRepo.GetQuestData(ctx)
	if err != nil {
		return err
	}

	for _, bucket := range questData.QuestMap {
		if bucket == nil {
			continue
		}
		for _, quest := range bucket.Quests {
			if quest == nil {
				continue
			}

			questConfig := uc.configManager.GetQuestConfig(quest.Id)
			if questConfig == nil {
				continue
			}

			for targetIndex, target := range questConfig.Targets {
				if target.Type != questType {
					continue
				}

				matched := false
				switch questType {
				case uint32(protocol.QuestType_QuestTypeTalkToNPC):
					if len(target.Ids) == 0 {
						matched = true
					} else {
						for _, id := range target.Ids {
							if id == targetId {
								matched = true
								break
							}
						}
					}
				case uint32(protocol.QuestType_QuestTypeLearnSkill):
					matched = true
				case uint32(protocol.QuestType_QuestTypeKillMonster):
					matched = true
				}

				if !matched {
					continue
				}

				for int(targetIndex) >= len(quest.Progress) {
					quest.Progress = append(quest.Progress, 0)
				}

				newProgress := quest.Progress[targetIndex] + count
				if newProgress > target.Count {
					newProgress = target.Count
				}
				quest.Progress[targetIndex] = newProgress

				log.Infof("Quest progress updated by type: RoleID=%d, QuestID=%d, TargetIndex=%d, Progress=%d/%d, Type=%d, TargetId=%d",
					roleID, quest.Id, targetIndex, newProgress, target.Count, questType, targetId)

				if uc.isQuestCompleted(questData, quest.Id) {
					log.Infof("Quest completed: RoleID=%d, QuestID=%d", roleID, quest.Id)
				}
			}
		}
	}

	return nil
}

// isQuestCompleted 检查任务是否完成
func (uc *UpdateQuestProgressUseCase) isQuestCompleted(questData *protocol.SiQuestData, questId uint32) bool {
	quest := uc.getQuest(questData, questId)
	if quest == nil {
		return false
	}

	questConfig := uc.configManager.GetQuestConfig(questId)
	if questConfig == nil {
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
func (uc *UpdateQuestProgressUseCase) getQuest(questData *protocol.SiQuestData, questId uint32) *protocol.QuestData {
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
