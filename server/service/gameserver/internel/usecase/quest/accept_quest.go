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

const (
	questCategoryMain   = uint32(protocol.QuestCategory_QuestCategoryMain)
	questCategoryBranch = uint32(protocol.QuestCategory_QuestCategoryBranch)
	questCategoryDaily  = uint32(protocol.QuestCategory_QuestCategoryDaily)
	questCategoryWeekly = uint32(protocol.QuestCategory_QuestCategoryWeekly)
)

// AcceptQuestUseCase 接受任务用例
type AcceptQuestUseCase struct {
	playerRepo    repository.PlayerRepository
	configManager interfaces.ConfigManager
	levelUseCase  interfaces.LevelUseCase
}

// NewAcceptQuestUseCase 创建接受任务用例
func NewAcceptQuestUseCase(
	playerRepo repository.PlayerRepository,
	configManager interfaces.ConfigManager,
) *AcceptQuestUseCase {
	return &AcceptQuestUseCase{
		playerRepo:    playerRepo,
		configManager: configManager,
	}
}

// SetDependencies 设置依赖（用于注入 LevelUseCase）
func (uc *AcceptQuestUseCase) SetDependencies(levelUseCase interfaces.LevelUseCase) {
	uc.levelUseCase = levelUseCase
}

// Execute 执行接受任务用例
func (uc *AcceptQuestUseCase) Execute(ctx context.Context, roleID uint64, questId uint32) error {
	// 获取任务配置
	questConfigRaw, ok := uc.configManager.GetQuestConfig(questId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "quest config not found: %d", questId)
	}

	questConfig, ok := questConfigRaw.(*jsonconf.QuestConfig)
	if !ok || questConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "invalid quest config type")
	}

	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	// 检查任务系统是否初始化
	if binaryData.QuestData == nil {
		binaryData.QuestData = &protocol.SiQuestData{
			QuestMap:     make(map[uint32]*protocol.QuestTypeList),
			LastResetMap: make(map[uint32]int64),
		}
	}
	if binaryData.QuestData.QuestMap == nil {
		binaryData.QuestData.QuestMap = make(map[uint32]*protocol.QuestTypeList)
	}

	// 检查是否已接受
	if uc.hasQuest(binaryData.QuestData, questId) {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "quest already accepted: %d", questId)
	}

	// 检查等级要求
	if questConfig.Level > 0 && uc.levelUseCase != nil {
		currentLevel, err := uc.levelUseCase.GetLevel(ctx, roleID)
		if err != nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "获取等级失败: %v", err)
		}
		if currentLevel < questConfig.Level {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "level not enough: need %d, current %d", questConfig.Level, currentLevel)
		}
	}

	// 检查前置任务
	if len(questConfig.PreQuests) > 0 {
		for _, preQuestId := range questConfig.PreQuests {
			preQuest := uc.getQuest(binaryData.QuestData, preQuestId)
			if preQuest == nil {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "pre quest not completed: %d", preQuestId)
			}
			// 检查前置任务是否完成
			if !uc.isQuestCompleted(binaryData.QuestData, preQuestId) {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "pre quest not completed: %d", preQuestId)
			}
		}
	}

	// 通过配置表中的类型字段确定任务所属的大类（主线/支线/日常/周常）
	questType := questConfig.Type
	bucket := uc.ensureBucket(binaryData.QuestData, questType)
	if bucket == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "quest bucket init failed: %d", questType)
	}

	questData := &protocol.QuestData{
		Id:       questId,
		Progress: make([]uint32, len(questConfig.Targets)),
	}

	// 初始化进度为0
	for i := range questData.Progress {
		questData.Progress[i] = 0
	}

	bucket.Quests = append(bucket.Quests, questData)

	log.Infof("Quest accepted: RoleID=%d, QuestID=%d, Type=%d", roleID, questId, questType)
	return nil
}

// hasQuest 检查是否已接受任务
func (uc *AcceptQuestUseCase) hasQuest(questData *protocol.SiQuestData, questId uint32) bool {
	return uc.getQuest(questData, questId) != nil
}

// getQuest 获取指定任务
func (uc *AcceptQuestUseCase) getQuest(questData *protocol.SiQuestData, questId uint32) *protocol.QuestData {
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

// isQuestCompleted 检查任务是否完成
func (uc *AcceptQuestUseCase) isQuestCompleted(questData *protocol.SiQuestData, questId uint32) bool {
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

// ensureBucket 确保任务桶存在
func (uc *AcceptQuestUseCase) ensureBucket(questData *protocol.SiQuestData, questType uint32) *protocol.QuestTypeList {
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
