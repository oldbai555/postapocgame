package quest

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
)

const (
	questCategoryMain   = uint32(protocol.QuestCategory_QuestCategoryMain)
	questCategoryBranch = uint32(protocol.QuestCategory_QuestCategoryBranch)
	questCategoryDaily  = uint32(protocol.QuestCategory_QuestCategoryDaily)
	questCategoryWeekly = uint32(protocol.QuestCategory_QuestCategoryWeekly)
)

// InitQuestDataUseCase 初始化任务数据用例
// 负责任务数据的初始化（任务桶结构、基础任务类型集合、可重复任务补齐）
type InitQuestDataUseCase struct {
	playerRepo    repository.PlayerRepository
	configManager interfaces.ConfigManager
}

// NewInitQuestDataUseCase 创建初始化任务数据用例
func NewInitQuestDataUseCase(
	playerRepo repository.PlayerRepository,
	configManager interfaces.ConfigManager,
) *InitQuestDataUseCase {
	return &InitQuestDataUseCase{
		playerRepo:    playerRepo,
		configManager: configManager,
	}
}

// Execute 执行初始化任务数据用例
func (uc *InitQuestDataUseCase) Execute(ctx context.Context, roleID uint64) error {
	// 获取 BinaryData（共享引用）
	questData, err := uc.playerRepo.GetQuestData(ctx)
	if err != nil {
		return err
	}

	// 初始化基础任务桶（主线/支线/日常/周常）
	uc.ensureBucket(questData, questCategoryMain)
	uc.ensureBucket(questData, questCategoryBranch)
	uc.ensureBucket(questData, questCategoryDaily)
	uc.ensureBucket(questData, questCategoryWeekly)

	return nil
}

// ensureBucket 确保任务桶存在
func (uc *InitQuestDataUseCase) ensureBucket(questData *protocol.SiQuestData, questType uint32) *protocol.QuestTypeList {
	if questData == nil {
		return nil
	}
	if questData.QuestMap == nil {
		questData.QuestMap = make(map[uint32]*protocol.QuestTypeList)
	}
	bucket, ok := questData.QuestMap[questType]
	if !ok || bucket == nil {
		bucket = &protocol.QuestTypeList{
			Quests: make([]*protocol.QuestData, 0),
		}
		questData.QuestMap[questType] = bucket
	}
	if bucket.Quests == nil {
		bucket.Quests = make([]*protocol.QuestData, 0)
	}
	return bucket
}
