package quest

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
	"time"
)

// RefreshQuestTypeUseCase 刷新任务类型用例
// 负责根据任务类型（每日/每周）刷新任务列表，包括：
// - 根据玩家等级筛选可用任务
// - 从配置中生成任务数据
// - 更新最后刷新时间
// - 清空对应类型的任务完成次数
type RefreshQuestTypeUseCase struct {
	playerRepo    repository.PlayerRepository
	configManager interfaces.ConfigManager
	levelUseCase  interfaces.LevelUseCase
}

// NewRefreshQuestTypeUseCase 创建刷新任务类型用例
func NewRefreshQuestTypeUseCase(
	playerRepo repository.PlayerRepository,
	configManager interfaces.ConfigManager,
) *RefreshQuestTypeUseCase {
	return &RefreshQuestTypeUseCase{
		playerRepo:    playerRepo,
		configManager: configManager,
	}
}

// SetDependencies 设置依赖（用于注入 LevelUseCase）
func (uc *RefreshQuestTypeUseCase) SetDependencies(levelUseCase interfaces.LevelUseCase) {
	uc.levelUseCase = levelUseCase
}

// Execute 执行刷新任务类型用例
// questType: 任务类型（questCategoryDaily 或 questCategoryWeekly）
func (uc *RefreshQuestTypeUseCase) Execute(ctx context.Context, roleID uint64, questType uint32) error {
	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	if binaryData.QuestData == nil {
		// 如果 QuestData 不存在，先初始化
		binaryData.QuestData = &protocol.SiQuestData{
			QuestMap:     make(map[uint32]*protocol.QuestTypeList),
			LastResetMap: make(map[uint32]int64),
		}
	}

	// 确保任务桶存在
	bucket := uc.ensureBucket(binaryData.QuestData, questType)
	if bucket == nil {
		return nil
	}

	// 获取玩家等级
	var level uint32
	if uc.levelUseCase != nil {
		level, _ = uc.levelUseCase.GetLevel(ctx, roleID)
	}

	// 从配置中获取该类型的所有任务配置
	configsRaw := uc.configManager.GetQuestConfigsByType(questType)
	bucket.Quests = bucket.Quests[:0]
	now := servertime.Now().Unix()

	// 根据玩家等级筛选并生成任务数据
	for _, cfgRaw := range configsRaw {
		if cfgRaw == nil {
			continue
		}
		cfg, ok := cfgRaw.(*jsonconf.QuestConfig)
		if !ok || cfg == nil {
			continue
		}
		// 只添加玩家等级满足要求的任务
		if cfg.Level > level {
			continue
		}
		bucket.Quests = append(bucket.Quests, uc.newQuestDataFromConfig(cfg))
	}

	// 更新最后刷新时间
	if binaryData.QuestData.LastResetMap == nil {
		binaryData.QuestData.LastResetMap = make(map[uint32]int64)
	}
	binaryData.QuestData.LastResetMap[questType] = now

	// 刷新该类型任务时，清空对应任务的完成次数，保证每日/每周统计从 0 开始
	if binaryData.QuestData.QuestFinishCount != nil {
		for questId := range binaryData.QuestData.QuestFinishCount {
			cfgRaw, ok := uc.configManager.GetQuestConfig(questId)
			if !ok {
				// 配置已删除或异常的任务，直接清理计数
				delete(binaryData.QuestData.QuestFinishCount, questId)
				continue
			}
			cfg, ok := cfgRaw.(*jsonconf.QuestConfig)
			if ok && cfg != nil && cfg.Type == questType {
				delete(binaryData.QuestData.QuestFinishCount, questId)
			}
		}
	}

	log.Infof("Quest type refreshed: RoleID=%d, Type=%d, Count=%d", roleID, questType, len(bucket.Quests))
	return nil
}

// ShouldRefresh 检查是否应该刷新（业务逻辑：判断时间是否跨天/跨周）
func (uc *RefreshQuestTypeUseCase) ShouldRefresh(questData *protocol.SiQuestData, questType uint32, now time.Time) bool {
	now = now.In(time.Local)
	if questData == nil || questData.LastResetMap == nil {
		return true
	}
	last := questData.LastResetMap[questType]
	if last == 0 {
		return true
	}
	lastTime := time.Unix(last, 0).In(time.Local)
	switch questType {
	case questCategoryDaily:
		return !isSameDay(now, lastTime)
	case questCategoryWeekly:
		return !isSameWeek(now, lastTime)
	default:
		return false
	}
}

// EnsureRepeatableQuests 确保可重复任务存在（在初始化时调用）
func (uc *RefreshQuestTypeUseCase) EnsureRepeatableQuests(ctx context.Context, roleID uint64, questData *protocol.SiQuestData) error {
	now := servertime.Now()
	if uc.ShouldRefresh(questData, questCategoryDaily, now) || len(uc.ensureBucket(questData, questCategoryDaily).Quests) == 0 {
		if err := uc.Execute(ctx, roleID, questCategoryDaily); err != nil {
			return err
		}
	}
	if uc.ShouldRefresh(questData, questCategoryWeekly, now) || len(uc.ensureBucket(questData, questCategoryWeekly).Quests) == 0 {
		if err := uc.Execute(ctx, roleID, questCategoryWeekly); err != nil {
			return err
		}
	}
	return nil
}

// ensureBucket 确保任务桶存在
func (uc *RefreshQuestTypeUseCase) ensureBucket(questData *protocol.SiQuestData, questType uint32) *protocol.QuestTypeList {
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

// newQuestDataFromConfig 从配置创建任务数据
func (uc *RefreshQuestTypeUseCase) newQuestDataFromConfig(cfg *jsonconf.QuestConfig) *protocol.QuestData {
	progress := make([]uint32, len(cfg.Targets))
	for i := range progress {
		progress[i] = 0
	}
	return &protocol.QuestData{
		Id:       cfg.QuestId,
		Progress: progress,
	}
}

// isSameDay 检查是否是同一天
func isSameDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// isSameWeek 检查是否是同一周
func isSameWeek(a, b time.Time) bool {
	y1, w1 := a.ISOWeek()
	y2, w2 := b.ISOWeek()
	return y1 == y2 && w1 == w2
}
