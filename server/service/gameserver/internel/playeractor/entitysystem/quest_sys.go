package entitysystem

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
)

// QuestSys 任务系统
type QuestSys struct {
	*BaseSystem
	questData *protocol.SiQuestData
}

// NewQuestSys 创建任务系统
func NewQuestSys() *QuestSys {
	return &QuestSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysQuest)),
	}
}

func GetQuestSys(ctx context.Context) *QuestSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysQuest))
	if system == nil {
		return nil
	}
	questSys, ok := system.(*QuestSys)
	if !ok || !questSys.IsOpened() {
		return nil
	}
	return questSys
}

// OnInit 系统初始化
func (qs *QuestSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("quest sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果quest_data不存在，则初始化
	if binaryData.QuestData == nil {
		binaryData.QuestData = &protocol.SiQuestData{
			Quests: make([]*protocol.QuestData, 0),
		}
	}
	qs.questData = binaryData.QuestData

	// 如果Quests为空，初始化为空切片
	if qs.questData.Quests == nil {
		qs.questData.Quests = make([]*protocol.QuestData, 0)
	}

	log.Infof("QuestSys initialized: QuestCount=%d", len(qs.questData.Quests))
}

// GetQuestData 获取任务数据
func (qs *QuestSys) GetQuestData() *protocol.SiQuestData {
	return qs.questData
}

// GetQuest 获取指定任务
func (qs *QuestSys) GetQuest(questId uint32) *protocol.QuestData {
	if qs.questData == nil || qs.questData.Quests == nil {
		return nil
	}
	for _, quest := range qs.questData.Quests {
		if quest != nil && quest.Id == questId {
			return quest
		}
	}
	return nil
}

// HasQuest 检查是否已接受任务
func (qs *QuestSys) HasQuest(questId uint32) bool {
	return qs.GetQuest(questId) != nil
}

// AcceptQuest 接受任务
func (qs *QuestSys) AcceptQuest(ctx context.Context, questId uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	// 检查任务配置
	questConfig, ok := jsonconf.GetConfigManager().GetQuestConfig(questId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "quest config not found: %d", questId)
	}

	// 检查是否已接受
	if qs.HasQuest(questId) {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "quest already accepted: %d", questId)
	}

	// 检查等级要求
	levelSys := GetLevelSys(ctx)
	if levelSys != nil {
		if levelSys.GetLevel() < questConfig.Level {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "level not enough: need %d, current %d", questConfig.Level, levelSys.GetLevel())
		}
	}

	// 检查前置任务
	if len(questConfig.PreQuests) > 0 {
		for _, preQuestId := range questConfig.PreQuests {
			preQuest := qs.GetQuest(preQuestId)
			if preQuest == nil {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "pre quest not completed: %d", preQuestId)
			}
			// 检查前置任务是否完成（所有目标进度都达到要求）
			if !qs.isQuestCompleted(preQuestId) {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "pre quest not completed: %d", preQuestId)
			}
		}
	}

	// 创建任务数据
	questData := &protocol.QuestData{
		Id:       questId,
		Progress: make([]uint32, len(questConfig.Targets)),
	}

	// 初始化进度为0
	for i := range questData.Progress {
		questData.Progress[i] = 0
	}

	// 添加到任务列表
	if qs.questData.Quests == nil {
		qs.questData.Quests = make([]*protocol.QuestData, 0)
	}
	qs.questData.Quests = append(qs.questData.Quests, questData)

	log.Infof("Quest accepted: RoleID=%d, QuestID=%d", playerRole.GetPlayerRoleId(), questId)
	return nil
}

// UpdateQuestProgress 更新任务进度（按目标索引）
func (qs *QuestSys) UpdateQuestProgress(ctx context.Context, questId uint32, targetIndex uint32, progress uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	// 获取任务
	quest := qs.GetQuest(questId)
	if quest == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "quest not found: %d", questId)
	}

	// 检查任务配置
	questConfig, ok := jsonconf.GetConfigManager().GetQuestConfig(questId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "quest config not found: %d", questId)
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

	log.Infof("Quest progress updated: RoleID=%d, QuestID=%d, TargetIndex=%d, Progress=%d/%d",
		playerRole.GetPlayerRoleId(), questId, targetIndex, progress, target.Count)

	// 检查任务是否完成
	if qs.isQuestCompleted(questId) {
		log.Infof("Quest completed: RoleID=%d, QuestID=%d", playerRole.GetPlayerRoleId(), questId)
	}

	return nil
}

// UpdateQuestProgressByType 根据任务类型更新进度（自动匹配符合条件的任务目标）
// questType: 任务类型（1=和NPC对话，2=学习技能，3=击杀怪物）
// targetId: 目标ID（对于type=1，传入npcId；对于type=2和3，可以传入0表示任意）
// count: 增加的数量（默认1）
func (qs *QuestSys) UpdateQuestProgressByType(ctx context.Context, questType uint32, targetId uint32, count uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	if qs.questData == nil || qs.questData.Quests == nil {
		return nil
	}

	// 遍历所有已接受的任务
	for _, quest := range qs.questData.Quests {
		if quest == nil {
			continue
		}

		questConfig, ok := jsonconf.GetConfigManager().GetQuestConfig(quest.Id)
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
					// ids为空，表示任意NPC都可以
					matched = true
				} else {
					// 检查targetId是否在ids中
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

				log.Infof("Quest progress updated by type: RoleID=%d, QuestID=%d, TargetIndex=%d, Progress=%d/%d, Type=%d, TargetId=%d",
					playerRole.GetPlayerRoleId(), quest.Id, targetIndex, newProgress, target.Count, questType, targetId)

				// 检查任务是否完成
				if qs.isQuestCompleted(quest.Id) {
					log.Infof("Quest completed: RoleID=%d, QuestID=%d", playerRole.GetPlayerRoleId(), quest.Id)
				}
			}
		}
	}

	return nil
}

// isQuestCompleted 检查任务是否完成
func (qs *QuestSys) isQuestCompleted(questId uint32) bool {
	quest := qs.GetQuest(questId)
	if quest == nil {
		return false
	}

	questConfig, ok := jsonconf.GetConfigManager().GetQuestConfig(questId)
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

// SubmitQuest 提交任务
func (qs *QuestSys) SubmitQuest(ctx context.Context, questId uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	// 获取任务
	quest := qs.GetQuest(questId)
	if quest == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "quest not found: %d", questId)
	}

	// 检查任务是否完成
	if !qs.isQuestCompleted(questId) {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "quest not completed: %d", questId)
	}

	// 获取任务配置
	questConfig, ok := jsonconf.GetConfigManager().GetQuestConfig(questId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "quest config not found: %d", questId)
	}

	// 发放经验奖励
	if questConfig.ExpReward > 0 {
		levelSys := GetLevelSys(ctx)
		if levelSys != nil {
			if err := levelSys.AddExp(ctx, questConfig.ExpReward); err != nil {
				log.Errorf("AddExp failed: %v", err)
				// 经验发放失败不影响任务提交，只记录日志
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
				Bind:     1, // 任务奖励默认绑定
			})
		}
		if err := playerRole.GrantRewards(ctx, rewards); err != nil {
			log.Errorf("GrantRewards failed: %v", err)
			return customerr.Wrap(err)
		}
	}

	// 从任务列表中移除
	qs.removeQuest(questId)

	// 检查是否有后续任务（任务链）
	if len(questConfig.NextQuests) > 0 {
		for _, nextQuestId := range questConfig.NextQuests {
			// 自动接取后续任务
			if err := qs.AcceptQuest(ctx, nextQuestId); err != nil {
				log.Warnf("Auto accept next quest failed: QuestID=%d, NextQuestID=%d, Error=%v", questId, nextQuestId, err)
			} else {
				log.Infof("Auto accepted next quest: QuestID=%d, NextQuestID=%d", questId, nextQuestId)
			}
		}
	}

	log.Infof("Quest submitted: RoleID=%d, QuestID=%d", playerRole.GetPlayerRoleId(), questId)
	return nil
}

// removeQuest 移除任务
func (qs *QuestSys) removeQuest(questId uint32) {
	if qs.questData == nil || qs.questData.Quests == nil {
		return
	}

	for i, quest := range qs.questData.Quests {
		if quest != nil && quest.Id == questId {
			// 从切片中移除
			qs.questData.Quests = append(qs.questData.Quests[:i], qs.questData.Quests[i+1:]...)
			return
		}
	}
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysQuest), func() iface.ISystem {
		return NewQuestSys()
	})
}
