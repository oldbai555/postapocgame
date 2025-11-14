package entitysystem

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
	"time"
)

// DailyQuestSys 每日任务系统
type DailyQuestSys struct {
	*BaseSystem
	dailyQuestData *protocol.SiDailyQuestData
}

// NewDailyQuestSys 创建每日任务系统
func NewDailyQuestSys() *DailyQuestSys {
	return &DailyQuestSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysDailyQuest)),
	}
}

func GetDailyQuestSys(ctx context.Context) *DailyQuestSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysDailyQuest))
	if system == nil {
		return nil
	}
	dailyQuestSys, ok := system.(*DailyQuestSys)
	if !ok || !dailyQuestSys.IsOpened() {
		return nil
	}
	return dailyQuestSys
}

// OnInit 系统初始化
func (dqs *DailyQuestSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("daily quest sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果daily_quest_data不存在，则初始化
	if binaryData.DailyQuestData == nil {
		binaryData.DailyQuestData = &protocol.SiDailyQuestData{
			DailyQuests: make([]*protocol.DailyQuestData, 0),
		}
	}
	dqs.dailyQuestData = binaryData.DailyQuestData

	// 如果DailyQuests为空，初始化为空切片
	if dqs.dailyQuestData.DailyQuests == nil {
		dqs.dailyQuestData.DailyQuests = make([]*protocol.DailyQuestData, 0)
	}

	// 检查是否需要刷新每日任务
	dqs.checkAndRefreshDailyQuests(ctx)

	log.Infof("DailyQuestSys initialized: QuestCount=%d", len(dqs.dailyQuestData.DailyQuests))
}

// checkAndRefreshDailyQuests 检查并刷新每日任务
func (dqs *DailyQuestSys) checkAndRefreshDailyQuests(ctx context.Context) {
	now := time.Now()
	lastRefreshTime := time.Unix(dqs.dailyQuestData.LastRefreshTime, 0)

	// 检查是否跨天（需要刷新）
	if now.Year() != lastRefreshTime.Year() || now.YearDay() != lastRefreshTime.YearDay() {
		// 清空旧的每日任务
		dqs.dailyQuestData.DailyQuests = make([]*protocol.DailyQuestData, 0)
		dqs.dailyQuestData.LastRefreshTime = now.Unix()

		// 自动刷新每日任务（从配置中获取所有每日任务）
		dqs.refreshDailyQuests(ctx)
	}
}

// refreshDailyQuests 刷新每日任务
func (dqs *DailyQuestSys) refreshDailyQuests(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("refreshDailyQuests get role err:%v", err)
		return
	}

	// 获取所有每日任务配置
	configManager := jsonconf.GetConfigManager()
	allDailyQuests := configManager.GetAllDailyQuestConfigs()

	// 根据玩家等级筛选可接受的每日任务
	levelSys := GetLevelSys(ctx)
	if levelSys == nil {
		log.Errorf("level sys not found")
		return
	}

	playerLevel := levelSys.GetLevel()
	now := time.Now().Unix()

	for _, questConfig := range allDailyQuests {
		// 检查等级要求
		if questConfig.Level > playerLevel {
			continue
		}

		// 创建每日任务数据
		dailyQuest := &protocol.DailyQuestData{
			QuestId:    questConfig.QuestId,
			Progress:   make([]uint32, len(questConfig.Targets)),
			AcceptTime: now,
		}

		// 初始化进度为0
		for i := range dailyQuest.Progress {
			dailyQuest.Progress[i] = 0
		}

		// 添加到每日任务列表
		dqs.dailyQuestData.DailyQuests = append(dqs.dailyQuestData.DailyQuests, dailyQuest)
	}

	log.Infof("Daily quests refreshed: RoleID=%d, Count=%d", playerRole.GetPlayerRoleId(), len(dqs.dailyQuestData.DailyQuests))
}

// GetDailyQuest 获取指定每日任务
func (dqs *DailyQuestSys) GetDailyQuest(questId uint32) *protocol.DailyQuestData {
	if dqs.dailyQuestData == nil || dqs.dailyQuestData.DailyQuests == nil {
		return nil
	}
	for _, quest := range dqs.dailyQuestData.DailyQuests {
		if quest != nil && quest.QuestId == questId {
			return quest
		}
	}
	return nil
}

// UpdateDailyQuestProgress 更新每日任务进度（按目标索引）
func (dqs *DailyQuestSys) UpdateDailyQuestProgress(ctx context.Context, questId uint32, targetIndex uint32, progress uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	// 获取每日任务
	quest := dqs.GetDailyQuest(questId)
	if quest == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "daily quest not found: %d", questId)
	}

	// 检查每日任务配置
	questConfig, ok := jsonconf.GetConfigManager().GetDailyQuestConfig(questId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "daily quest config not found: %d", questId)
	}

	// 检查目标索引
	if int(targetIndex) >= len(questConfig.Targets) {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "invalid target index: %d", targetIndex)
	}

	// 更新进度（不能超过目标数量）
	target := questConfig.Targets[targetIndex]
	if progress > target.Count {
		progress = target.Count
	}

	// 确保Progress数组足够大
	for int(targetIndex) >= len(quest.Progress) {
		quest.Progress = append(quest.Progress, 0)
	}

	quest.Progress[targetIndex] = progress

	log.Infof("Daily quest progress updated: RoleID=%d, QuestID=%d, TargetIndex=%d, Progress=%d/%d",
		playerRole.GetPlayerRoleId(), questId, targetIndex, progress, target.Count)

	// 检查任务是否完成
	if dqs.isDailyQuestCompleted(questId) {
		log.Infof("Daily quest completed: RoleID=%d, QuestID=%d", playerRole.GetPlayerRoleId(), questId)
	}

	return nil
}

// UpdateDailyQuestProgressByType 根据任务类型更新进度（自动匹配符合条件的任务目标）
func (dqs *DailyQuestSys) UpdateDailyQuestProgressByType(ctx context.Context, questType uint32, targetId uint32, count uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	if dqs.dailyQuestData == nil || dqs.dailyQuestData.DailyQuests == nil {
		return nil
	}

	// 遍历所有每日任务
	for _, quest := range dqs.dailyQuestData.DailyQuests {
		if quest == nil {
			continue
		}

		questConfig, ok := jsonconf.GetConfigManager().GetDailyQuestConfig(quest.QuestId)
		if !ok {
			continue
		}

		// 遍历任务的所有目标
		for targetIndex, target := range questConfig.Targets {
			// 检查任务类型是否匹配
			if target.Type != questType {
				continue
			}

			// 根据任务类型进行匹配检查
			matched := false
			switch questType {
			case uint32(protocol.QuestType_QuestTypeTalkToNPC):
				// type=1: 和NPC对话，需要检查ids中是否包含targetId
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
				// type=2: 学习任意技能，ids不需要配置
				matched = true
			case uint32(protocol.QuestType_QuestTypeKillMonster):
				// type=3: 击杀任意怪物，ids不需要配置
				matched = true
			}

			if matched {
				// 确保Progress数组足够大
				for int(targetIndex) >= len(quest.Progress) {
					quest.Progress = append(quest.Progress, 0)
				}

				// 增加进度（不能超过目标数量）
				newProgress := quest.Progress[targetIndex] + count
				if newProgress > target.Count {
					newProgress = target.Count
				}
				quest.Progress[targetIndex] = newProgress

				log.Infof("Daily quest progress updated by type: RoleID=%d, QuestID=%d, TargetIndex=%d, Progress=%d/%d, Type=%d, TargetId=%d",
					playerRole.GetPlayerRoleId(), quest.QuestId, targetIndex, newProgress, target.Count, questType, targetId)

				// 检查任务是否完成
				if dqs.isDailyQuestCompleted(quest.QuestId) {
					log.Infof("Daily quest completed: RoleID=%d, QuestID=%d", playerRole.GetPlayerRoleId(), quest.QuestId)
				}
			}
		}
	}

	return nil
}

// isDailyQuestCompleted 检查每日任务是否完成
func (dqs *DailyQuestSys) isDailyQuestCompleted(questId uint32) bool {
	quest := dqs.GetDailyQuest(questId)
	if quest == nil {
		return false
	}

	questConfig, ok := jsonconf.GetConfigManager().GetDailyQuestConfig(questId)
	if !ok {
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

// SubmitDailyQuest 提交每日任务
func (dqs *DailyQuestSys) SubmitDailyQuest(ctx context.Context, questId uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	// 获取每日任务
	quest := dqs.GetDailyQuest(questId)
	if quest == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "daily quest not found: %d", questId)
	}

	// 检查任务是否完成
	if !dqs.isDailyQuestCompleted(questId) {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "daily quest not completed: %d", questId)
	}

	// 获取每日任务配置
	questConfig, ok := jsonconf.GetConfigManager().GetDailyQuestConfig(questId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "daily quest config not found: %d", questId)
	}

	// 发放经验奖励
	if questConfig.ExpReward > 0 {
		levelSys := GetLevelSys(ctx)
		if levelSys != nil {
			if err := levelSys.AddExp(ctx, questConfig.ExpReward); err != nil {
				log.Errorf("AddExp failed: %v", err)
			}
		}
	}

	// 发放物品奖励
	if len(questConfig.Rewards) > 0 {
		rewards := make([]*jsonconf.ItemAmount, 0, len(questConfig.Rewards))
		for _, reward := range questConfig.Rewards {
			rewards = append(rewards, &jsonconf.ItemAmount{
				ItemType: uint32(reward.Type),
				ItemId:   reward.ItemId,
				Count:    int64(reward.Count),
				Bind:     1, // 每日任务奖励默认绑定
			})
		}
		if err := playerRole.GrantRewards(ctx, rewards); err != nil {
			log.Errorf("GrantRewards failed: %v", err)
			return customerr.Wrap(err)
		}
	}

	// 重置任务进度（每日任务可以重复完成）
	for i := range quest.Progress {
		quest.Progress[i] = 0
	}
	quest.AcceptTime = time.Now().Unix()

	log.Infof("Daily quest submitted: RoleID=%d, QuestID=%d", playerRole.GetPlayerRoleId(), questId)
	return nil
}

// GetDailyQuestData 获取每日任务数据
func (dqs *DailyQuestSys) GetDailyQuestData() *protocol.SiDailyQuestData {
	return dqs.dailyQuestData
}

// OnRoleLogin 玩家登录时检查刷新
func (dqs *DailyQuestSys) OnRoleLogin(ctx context.Context) {
	dqs.checkAndRefreshDailyQuests(ctx)
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysDailyQuest), func() iface.ISystem {
		return NewDailyQuestSys()
	})
}
