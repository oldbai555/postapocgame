package system

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/usecaseadapter"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/quest"
	"time"
)

const (
	questCategoryMain   = uint32(protocol.QuestCategory_QuestCategoryMain)
	questCategoryBranch = uint32(protocol.QuestCategory_QuestCategoryBranch)
	questCategoryDaily  = uint32(protocol.QuestCategory_QuestCategoryDaily)
	questCategoryWeekly = uint32(protocol.QuestCategory_QuestCategoryWeekly)
)

// QuestSystemAdapter 任务系统适配器
type QuestSystemAdapter struct {
	*BaseSystemAdapter
	acceptQuestUseCase    *quest.AcceptQuestUseCase
	updateProgressUseCase *quest.UpdateQuestProgressUseCase
	submitQuestUseCase    *quest.SubmitQuestUseCase
}

// NewQuestSystemAdapter 创建任务系统适配器
func NewQuestSystemAdapter() *QuestSystemAdapter {
	container := di.GetContainer()
	acceptQuestUC := quest.NewAcceptQuestUseCase(container.PlayerGateway(), container.ConfigGateway())
	updateProgressUC := quest.NewUpdateQuestProgressUseCase(container.PlayerGateway(), container.ConfigGateway())
	submitQuestUC := quest.NewSubmitQuestUseCase(container.PlayerGateway(), container.ConfigGateway())

	// 注入依赖
	levelUseCase := NewLevelUseCaseAdapter()
	rewardUseCase := usecaseadapter.NewRewardUseCaseAdapter()
	dailyActivityUseCase := NewDailyActivityUseCaseAdapter()
	acceptQuestUC.SetDependencies(levelUseCase)
	submitQuestUC.SetDependencies(levelUseCase, rewardUseCase, dailyActivityUseCase)

	return &QuestSystemAdapter{
		BaseSystemAdapter:     NewBaseSystemAdapter(uint32(protocol.SystemId_SysQuest)),
		acceptQuestUseCase:    acceptQuestUC,
		updateProgressUseCase: updateProgressUC,
		submitQuestUseCase:    submitQuestUC,
	}
}

// OnInit 系统初始化
func (a *QuestSystemAdapter) OnInit(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("quest sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		log.Errorf("quest sys OnInit get binary data err:%v", err)
		return
	}

	if binaryData.QuestData == nil {
		binaryData.QuestData = &protocol.SiQuestData{
			QuestMap:     make(map[uint32]*protocol.QuestTypeList),
			LastResetMap: make(map[uint32]int64),
		}
	}
	if binaryData.QuestData.QuestMap == nil {
		binaryData.QuestData.QuestMap = make(map[uint32]*protocol.QuestTypeList)
	}
	if binaryData.QuestData.LastResetMap == nil {
		binaryData.QuestData.LastResetMap = make(map[uint32]int64)
	}

	// 初始化基础任务桶
	a.ensureBucket(binaryData.QuestData, questCategoryMain)
	a.ensureBucket(binaryData.QuestData, questCategoryBranch)
	a.ensureBucket(binaryData.QuestData, questCategoryDaily)
	a.ensureBucket(binaryData.QuestData, questCategoryWeekly)

	// 确保可重复任务存在
	a.ensureRepeatableQuests(ctx, binaryData.QuestData)

	log.Infof("QuestSys initialized: RoleID=%d, QuestTypeCount=%d", roleID, len(binaryData.QuestData.QuestMap))
}

// AcceptQuest 接受任务（对外接口，供其他系统调用）
func (a *QuestSystemAdapter) AcceptQuest(ctx context.Context, questId uint32) error {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	return a.acceptQuestUseCase.Execute(ctx, roleID, questId)
}

// UpdateQuestProgressByType 根据任务类型更新进度（对外接口，供其他系统调用）
func (a *QuestSystemAdapter) UpdateQuestProgressByType(ctx context.Context, questType uint32, targetId uint32, count uint32) error {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	return a.updateProgressUseCase.Execute(ctx, roleID, questType, targetId, count)
}

// SubmitQuest 提交任务（对外接口，供其他系统调用）
func (a *QuestSystemAdapter) SubmitQuest(ctx context.Context, questId uint32) error {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	return a.submitQuestUseCase.Execute(ctx, roleID, questId)
}

// GetQuest 获取指定任务
func (a *QuestSystemAdapter) GetQuest(ctx context.Context, questId uint32) (*protocol.QuestData, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return a.getQuest(binaryData.QuestData, questId), nil
}

// GetQuestData 获取任务数据（用于协议）
func (a *QuestSystemAdapter) GetQuestData(ctx context.Context) (*protocol.SiQuestData, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if binaryData.QuestData == nil {
		return &protocol.SiQuestData{
			QuestMap:     make(map[uint32]*protocol.QuestTypeList),
			LastResetMap: make(map[uint32]int64),
		}, nil
	}
	return binaryData.QuestData, nil
}

// OnNewDay 每日刷新（实现 ISystem 接口）
func (a *QuestSystemAdapter) OnNewDay(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("OnNewDay get role err:%v", err)
		return
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		log.Errorf("OnNewDay get binary data err:%v", err)
		return
	}
	a.refreshQuestType(ctx, binaryData.QuestData, questCategoryDaily)
}

// OnNewWeek 每周刷新（实现 ISystem 接口）
func (a *QuestSystemAdapter) OnNewWeek(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("OnNewWeek get role err:%v", err)
		return
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		log.Errorf("OnNewWeek get binary data err:%v", err)
		return
	}
	a.refreshQuestType(ctx, binaryData.QuestData, questCategoryWeekly)
}

// ensureBucket 确保任务桶存在
func (a *QuestSystemAdapter) ensureBucket(questData *protocol.SiQuestData, questType uint32) *protocol.QuestTypeList {
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

// ensureRepeatableQuests 确保可重复任务存在
func (a *QuestSystemAdapter) ensureRepeatableQuests(ctx context.Context, questData *protocol.SiQuestData) {
	now := servertime.Now()
	if a.shouldRefresh(questData, questCategoryDaily, now) || len(a.ensureBucket(questData, questCategoryDaily).Quests) == 0 {
		a.refreshQuestType(ctx, questData, questCategoryDaily)
	}
	if a.shouldRefresh(questData, questCategoryWeekly, now) || len(a.ensureBucket(questData, questCategoryWeekly).Quests) == 0 {
		a.refreshQuestType(ctx, questData, questCategoryWeekly)
	}
}

// refreshQuestType 刷新任务类型
func (a *QuestSystemAdapter) refreshQuestType(ctx context.Context, questData *protocol.SiQuestData, questType uint32) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("refreshQuestType get role err:%v", err)
		return
	}

	bucket := a.ensureBucket(questData, questType)
	if bucket == nil {
		return
	}

	levelUseCase := NewLevelUseCaseAdapter()
	var level uint32
	if levelUseCase != nil {
		level, _ = levelUseCase.GetLevel(ctx, roleID)
	}

	configsRaw := di.GetContainer().ConfigGateway().GetQuestConfigsByType(questType)
	bucket.Quests = bucket.Quests[:0]
	now := servertime.Now().Unix()

	for _, cfgRaw := range configsRaw {
		if cfgRaw == nil {
			continue
		}
		cfg, ok := cfgRaw.(*jsonconf.QuestConfig)
		if !ok || cfg == nil {
			continue
		}
		if cfg.Level > level {
			continue
		}
		bucket.Quests = append(bucket.Quests, a.newQuestDataFromConfig(cfg))
	}

	if questData.LastResetMap == nil {
		questData.LastResetMap = make(map[uint32]int64)
	}
	questData.LastResetMap[questType] = now

	// 刷新该类型任务时，清空对应任务的完成次数，保证每日/每周统计从 0 开始
	if questData.QuestFinishCount != nil {
		for questId := range questData.QuestFinishCount {
			cfgRaw, ok := di.GetContainer().ConfigGateway().GetQuestConfig(questId)
			if !ok {
				// 配置已删除或异常的任务，直接清理计数
				delete(questData.QuestFinishCount, questId)
				continue
			}
			cfg, ok := cfgRaw.(*jsonconf.QuestConfig)
			if ok && cfg != nil && cfg.Type == questType {
				delete(questData.QuestFinishCount, questId)
			}
		}
	}

	log.Infof("Quest type refreshed: RoleID=%d, Type=%d, Count=%d", roleID, questType, len(bucket.Quests))
}

// newQuestDataFromConfig 从配置创建任务数据
func (a *QuestSystemAdapter) newQuestDataFromConfig(cfg *jsonconf.QuestConfig) *protocol.QuestData {
	progress := make([]uint32, len(cfg.Targets))
	for i := range progress {
		progress[i] = 0
	}
	return &protocol.QuestData{
		Id:       cfg.QuestId,
		Progress: progress,
	}
}

// shouldRefresh 检查是否应该刷新
func (a *QuestSystemAdapter) shouldRefresh(questData *protocol.SiQuestData, questType uint32, now time.Time) bool {
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

// getQuest 获取指定任务
func (a *QuestSystemAdapter) getQuest(questData *protocol.SiQuestData, questId uint32) *protocol.QuestData {
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

// EnsureISystem 确保 QuestSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*QuestSystemAdapter)(nil)
