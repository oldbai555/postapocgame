package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/usecaseadapter"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/quest"
)

const (
	questCategoryDaily  = uint32(protocol.QuestCategory_QuestCategoryDaily)
	questCategoryWeekly = uint32(protocol.QuestCategory_QuestCategoryWeekly)
)

// QuestSystemAdapter 任务系统适配器
//
// 生命周期职责：
// - OnInit: 调用 InitQuestDataUseCase 初始化任务数据结构，调用 RefreshQuestTypeUseCase 确保可重复任务存在
// - OnNewDay: 调用 RefreshQuestTypeUseCase 刷新每日任务
// - OnNewWeek: 调用 RefreshQuestTypeUseCase 刷新每周任务
// - 其他生命周期: 暂未使用
//
// 业务逻辑：所有业务逻辑（任务接取/更新/提交/刷新）均在 UseCase 层实现
//
// ⚠️ 防退化机制：禁止在 SystemAdapter 中编写业务规则逻辑，只允许调用 UseCase 与管理生命周期
type QuestSystemAdapter struct {
	*BaseSystemAdapter
	acceptQuestUseCase      *quest.AcceptQuestUseCase
	updateProgressUseCase   *quest.UpdateQuestProgressUseCase
	submitQuestUseCase      *quest.SubmitQuestUseCase
	initQuestDataUseCase    *quest.InitQuestDataUseCase
	refreshQuestTypeUseCase *quest.RefreshQuestTypeUseCase
}

// NewQuestSystemAdapter 创建任务系统适配器
func NewQuestSystemAdapter() *QuestSystemAdapter {
	container := di.GetContainer()
	acceptQuestUC := quest.NewAcceptQuestUseCase(container.PlayerGateway(), container.ConfigGateway())
	updateProgressUC := quest.NewUpdateQuestProgressUseCase(container.PlayerGateway(), container.ConfigGateway())
	submitQuestUC := quest.NewSubmitQuestUseCase(container.PlayerGateway(), container.ConfigGateway())
	initQuestDataUC := quest.NewInitQuestDataUseCase(container.PlayerGateway(), container.ConfigGateway())
	refreshQuestTypeUC := quest.NewRefreshQuestTypeUseCase(container.PlayerGateway(), container.ConfigGateway())

	// 注入依赖
	levelUseCase := NewLevelUseCaseAdapter()
	rewardUseCase := usecaseadapter.NewRewardUseCaseAdapter()
	acceptQuestUC.SetDependencies(levelUseCase)
	submitQuestUC.SetDependencies(levelUseCase, rewardUseCase)
	refreshQuestTypeUC.SetDependencies(levelUseCase)

	return &QuestSystemAdapter{
		BaseSystemAdapter:       NewBaseSystemAdapter(uint32(protocol.SystemId_SysQuest)),
		acceptQuestUseCase:      acceptQuestUC,
		updateProgressUseCase:   updateProgressUC,
		submitQuestUseCase:      submitQuestUC,
		initQuestDataUseCase:    initQuestDataUC,
		refreshQuestTypeUseCase: refreshQuestTypeUC,
	}
}

// OnInit 系统初始化
func (a *QuestSystemAdapter) OnInit(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("quest sys OnInit get role err:%v", err)
		return
	}
	// 初始化任务数据（包括任务桶结构、基础任务类型集合等业务逻辑）
	if err := a.initQuestDataUseCase.Execute(ctx, roleID); err != nil {
		log.Errorf("quest sys OnInit init quest data err:%v", err)
		return
	}
	// 确保可重复任务存在（调用 UseCase 处理刷新逻辑）
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err == nil && binaryData != nil && binaryData.QuestData != nil {
		if err := a.refreshQuestTypeUseCase.EnsureRepeatableQuests(ctx, roleID, binaryData.QuestData); err != nil {
			log.Errorf("quest sys OnInit ensure repeatable quests err:%v", err)
		}
	}
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
	// 调用 UseCase 刷新每日任务（业务逻辑已下沉到 UseCase）
	if err := a.refreshQuestTypeUseCase.Execute(ctx, roleID, questCategoryDaily); err != nil {
		log.Errorf("OnNewDay refresh quest type err:%v", err)
	}
}

// OnNewWeek 每周刷新（实现 ISystem 接口）
func (a *QuestSystemAdapter) OnNewWeek(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("OnNewWeek get role err:%v", err)
		return
	}
	// 调用 UseCase 刷新每周任务（业务逻辑已下沉到 UseCase）
	if err := a.refreshQuestTypeUseCase.Execute(ctx, roleID, questCategoryWeekly); err != nil {
		log.Errorf("OnNewWeek refresh quest type err:%v", err)
	}
}

// 注意：以下方法已移除，业务逻辑已下沉到 RefreshQuestTypeUseCase
// - ensureBucket
// - ensureRepeatableQuests
// - refreshQuestType
// - newQuestDataFromConfig
// - shouldRefresh

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

// EnsureISystem 确保 QuestSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*QuestSystemAdapter)(nil)
