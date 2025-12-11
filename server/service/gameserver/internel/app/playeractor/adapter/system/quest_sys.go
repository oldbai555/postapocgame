package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/quest"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/reward"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

const (
	questCategoryDaily  = uint32(protocol.QuestCategory_QuestCategoryDaily)
	questCategoryWeekly = uint32(protocol.QuestCategory_QuestCategoryWeekly)
)

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
	acceptQuestUC := quest.NewAcceptQuestUseCase(deps.PlayerGateway(), deps.ConfigGateway())
	updateProgressUC := quest.NewUpdateQuestProgressUseCase(deps.PlayerGateway(), deps.ConfigGateway())
	submitQuestUC := quest.NewSubmitQuestUseCase(deps.PlayerGateway(), deps.ConfigGateway())
	initQuestDataUC := quest.NewInitQuestDataUseCase(deps.PlayerGateway(), deps.ConfigGateway())
	refreshQuestTypeUC := quest.NewRefreshQuestTypeUseCase(deps.PlayerGateway(), deps.ConfigGateway())

	// 注入依赖
	levelUseCase := NewLevelUseCaseAdapter()
	rewardUseCase := reward.NewRewardUseCase(deps.PlayerGateway(), deps.EventPublisher(), deps.ConfigGateway())
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
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("quest sys OnInit get role err:%v", err)
		return
	}
	// 初始化任务数据（包括任务桶结构、基础任务类型集合等业务逻辑）
	if err := a.initQuestDataUseCase.Execute(ctx, roleID); err != nil {
		log.Errorf("quest sys OnInit init quest data err:%v", err)
		return
	}
}

// AcceptQuest 接受任务（对外接口，供其他系统调用）
func (a *QuestSystemAdapter) AcceptQuest(ctx context.Context, questId uint32) error {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	return a.acceptQuestUseCase.Execute(ctx, roleID, questId)
}

// UpdateQuestProgressByType 根据任务类型更新进度（对外接口，供其他系统调用）
func (a *QuestSystemAdapter) UpdateQuestProgressByType(ctx context.Context, questType uint32, targetId uint32, count uint32) error {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	return a.updateProgressUseCase.Execute(ctx, roleID, questType, targetId, count)
}

// SubmitQuest 提交任务（对外接口，供其他系统调用）
func (a *QuestSystemAdapter) SubmitQuest(ctx context.Context, questId uint32) error {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	return a.submitQuestUseCase.Execute(ctx, roleID, questId)
}

// GetQuest 获取指定任务
func (a *QuestSystemAdapter) GetQuest(ctx context.Context, questId uint32) (*protocol.QuestData, error) {
	questData, err := deps.PlayerGateway().GetQuestData(ctx)
	if err != nil {
		return nil, err
	}
	return a.getQuest(questData, questId), nil
}

// GetQuestData 获取任务数据（用于协议）
func (a *QuestSystemAdapter) GetQuestData(ctx context.Context) (*protocol.SiQuestData, error) {
	questData, err := deps.PlayerGateway().GetQuestData(ctx)
	if err != nil {
		return nil, err
	}
	return questData, nil
}

// OnNewDay 每日刷新（实现 ISystem 接口）
func (a *QuestSystemAdapter) OnNewDay(ctx context.Context) {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
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
	roleID, err := gshare.GetRoleIDFromContext(ctx)
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

// GetQuestSys 获取任务系统
func GetQuestSys(ctx context.Context) *QuestSystemAdapter {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysQuest))
	if system == nil {
		return nil
	}
	sys, ok := system.(*QuestSystemAdapter)
	if !ok || !sys.IsOpened() {
		return nil
	}
	return sys
}

// 注册系统工厂和协议
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysQuest), func() iface.ISystem {
		return NewQuestSystemAdapter()
	})

	// 每日/每周刷新由时间事件统一驱动，这里暂不直接订阅 gevent
}
