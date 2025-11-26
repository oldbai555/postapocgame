package entitysystem

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/manager"
	"postapocgame/server/service/gameserver/internel/playeractor/clientprotocol"
	"time"
)

// QuestSys 任务系统
type QuestSys struct {
	*BaseSystem
	questData *protocol.SiQuestData
}

const (
	questCategoryMain   = uint32(protocol.QuestCategory_QuestCategoryMain)
	questCategoryBranch = uint32(protocol.QuestCategory_QuestCategoryBranch)
	questCategoryDaily  = uint32(protocol.QuestCategory_QuestCategoryDaily)
	questCategoryWeekly = uint32(protocol.QuestCategory_QuestCategoryWeekly)
)

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

	if binaryData.QuestData == nil {
		binaryData.QuestData = &protocol.SiQuestData{
			QuestMap:     make(map[uint32]*protocol.QuestTypeList),
			LastResetMap: make(map[uint32]int64),
		}
	}
	qs.questData = binaryData.QuestData
	if qs.questData.QuestMap == nil {
		qs.questData.QuestMap = make(map[uint32]*protocol.QuestTypeList)
	}
	if qs.questData.LastResetMap == nil {
		qs.questData.LastResetMap = make(map[uint32]int64)
	}

	// 初始化基础任务桶
	qs.ensureBucket(questCategoryMain)
	qs.ensureBucket(questCategoryBranch)
	qs.ensureBucket(questCategoryDaily)
	qs.ensureBucket(questCategoryWeekly)

	qs.ensureRepeatableQuests(ctx)

	log.Infof("QuestSys initialized: RoleID=%d, QuestTypeCount=%d", playerRole.GetPlayerRoleId(), len(qs.questData.QuestMap))
}

// GetQuestData 获取任务数据
func (qs *QuestSys) GetQuestData() *protocol.SiQuestData {
	return qs.questData
}

// GetQuest 获取指定任务
func (qs *QuestSys) GetQuest(questId uint32) *protocol.QuestData {
	quest, _ := qs.getQuestWithType(questId)
	return quest
}

func (qs *QuestSys) ensureBucket(questType uint32) *protocol.QuestTypeList {
	if qs.questData == nil {
		return nil
	}
	if qs.questData.QuestMap == nil {
		qs.questData.QuestMap = make(map[uint32]*protocol.QuestTypeList)
	}
	bucket, ok := qs.questData.QuestMap[questType]
	if !ok || bucket == nil {
		bucket = &protocol.QuestTypeList{
			Quests: make([]*protocol.QuestData, 0),
		}
		qs.questData.QuestMap[questType] = bucket
	}
	if bucket.Quests == nil {
		bucket.Quests = make([]*protocol.QuestData, 0)
	}
	return bucket
}

func (qs *QuestSys) getQuestWithType(questId uint32) (*protocol.QuestData, uint32) {
	if qs.questData == nil || qs.questData.QuestMap == nil {
		return nil, 0
	}
	for questType, bucket := range qs.questData.QuestMap {
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
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "quest config not found: %d (role=%d)", questId, playerRole.GetPlayerRoleId())
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

	// 通过配置表中的类型字段确定任务所属的大类（主线/支线/日常/周常）
	questType := questConfig.Type
	bucket := qs.ensureBucket(questType)
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

	log.Infof("Quest accepted: RoleID=%d, QuestID=%d, Type=%d", playerRole.GetPlayerRoleId(), questId, questType)
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

	if qs.questData == nil || qs.questData.QuestMap == nil {
		return nil
	}

	for _, bucket := range qs.questData.QuestMap {
		if bucket == nil {
			continue
		}
		for _, quest := range bucket.Quests {
			if quest == nil {
				continue
			}

			questConfig, ok := jsonconf.GetConfigManager().GetQuestConfig(quest.Id)
			if !ok {
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
					playerRole.GetPlayerRoleId(), quest.Id, targetIndex, newProgress, target.Count, questType, targetId)

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
	quest, questType := qs.getQuestWithType(questId)
	if quest == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "quest not found: %d (role=%d)", questId, playerRole.GetPlayerRoleId())
	}

	// 检查任务是否完成
	if !qs.isQuestCompleted(questId) {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "quest not completed: %d (role=%d)", questId, playerRole.GetPlayerRoleId())
	}

	// 获取任务配置
	questConfig, ok := jsonconf.GetConfigManager().GetQuestConfig(questId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "quest config not found: %d (role=%d)", questId, playerRole.GetPlayerRoleId())
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

	// 日常任务奖励活跃点
	if questType == questCategoryDaily && questConfig.ActivePoint > 0 {
		if dailyActivity := GetDailyActivitySys(ctx); dailyActivity != nil {
			if err := dailyActivity.AddActivePoints(ctx, questConfig.ActivePoint); err != nil {
				log.Warnf("AddActivePoints failed: RoleID=%d, QuestID=%d, Err=%v", playerRole.GetPlayerRoleId(), questId, err)
			}
		}
	}

	switch questType {
	case questCategoryDaily, questCategoryWeekly:
		// 日常/周常任务：支持在同一自然日/周内多次完成，受 MaxCount 限制
		if qs.questData.QuestFinishCount == nil {
			qs.questData.QuestFinishCount = make(map[uint32]uint32)
		}
		finishCount := qs.questData.QuestFinishCount[questId] + 1
		qs.questData.QuestFinishCount[questId] = finishCount

		if questConfig.MaxCount > 0 && finishCount >= questConfig.MaxCount {
			// 达到最大完成次数后，从当前桶中移除，等待下一次刷新重新生成
			qs.removeQuest(questId)
		} else {
			// 未达到最大次数，重置进度，允许同一自然日/周内再次完成
			qs.resetQuestProgress(quest)
		}
	default:
		qs.removeQuest(questId)
	}

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
	if qs.questData == nil || qs.questData.QuestMap == nil {
		return
	}

	for questType, bucket := range qs.questData.QuestMap {
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

func (qs *QuestSys) resetQuestProgress(quest *protocol.QuestData) {
	if quest == nil {
		return
	}
	for i := range quest.Progress {
		quest.Progress[i] = 0
	}
}

func (qs *QuestSys) ensureRepeatableQuests(ctx context.Context) {
	now := servertime.Now()
	if qs.shouldRefresh(questCategoryDaily, now) || len(qs.ensureBucket(questCategoryDaily).Quests) == 0 {
		qs.refreshQuestType(ctx, questCategoryDaily)
	}
	if qs.shouldRefresh(questCategoryWeekly, now) || len(qs.ensureBucket(questCategoryWeekly).Quests) == 0 {
		qs.refreshQuestType(ctx, questCategoryWeekly)
	}
}

func (qs *QuestSys) refreshQuestType(ctx context.Context, questType uint32) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("refreshQuestType get role err:%v", err)
		return
	}

	bucket := qs.ensureBucket(questType)
	if bucket == nil {
		return
	}

	levelSys := GetLevelSys(ctx)
	var level uint32
	if levelSys != nil {
		level = levelSys.GetLevel()
	}

	configs := jsonconf.GetConfigManager().GetQuestConfigsByType(questType)
	bucket.Quests = bucket.Quests[:0]
	now := servertime.Now().Unix()

	for _, cfg := range configs {
		if cfg == nil {
			continue
		}
		if levelSys != nil && cfg.Level > level {
			continue
		}
		bucket.Quests = append(bucket.Quests, qs.newQuestDataFromConfig(cfg))
	}

	if qs.questData.LastResetMap == nil {
		qs.questData.LastResetMap = make(map[uint32]int64)
	}
	qs.questData.LastResetMap[questType] = now

	// 刷新该类型任务时，清空对应任务的完成次数，保证每日/每周统计从 0 开始
	if qs.questData.QuestFinishCount != nil {
		for questId := range qs.questData.QuestFinishCount {
			cfg, ok := jsonconf.GetConfigManager().GetQuestConfig(questId)
			if !ok {
				// 配置已删除或异常的任务，直接清理计数
				delete(qs.questData.QuestFinishCount, questId)
				continue
			}
			if cfg.Type == questType {
				delete(qs.questData.QuestFinishCount, questId)
			}
		}
	}

	log.Infof("Quest type refreshed: RoleID=%d, Type=%d, Count=%d", playerRole.GetPlayerRoleId(), questType, len(bucket.Quests))
}

func (qs *QuestSys) newQuestDataFromConfig(cfg *jsonconf.QuestConfig) *protocol.QuestData {
	progress := make([]uint32, len(cfg.Targets))
	for i := range progress {
		progress[i] = 0
	}
	return &protocol.QuestData{
		Id:       cfg.QuestId,
		Progress: progress,
	}
}

func (qs *QuestSys) shouldRefresh(questType uint32, now time.Time) bool {
	now = now.In(time.Local)
	if qs.questData == nil || qs.questData.LastResetMap == nil {
		return true
	}
	last := qs.questData.LastResetMap[questType]
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

func (qs *QuestSys) OnNewDay(ctx context.Context) {
	qs.refreshQuestType(ctx, questCategoryDaily)
}

func (qs *QuestSys) OnNewWeek(ctx context.Context) {
	qs.refreshQuestType(ctx, questCategoryWeekly)
}

func isSameDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func isSameWeek(a, b time.Time) bool {
	y1, w1 := a.ISOWeek()
	y2, w2 := b.ISOWeek()
	return y1 == y2 && w1 == w2
}

// handleTalkToNPC 处理NPC对话
func handleTalkToNPC(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	var req protocol.C2STalkToNPCReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// 获取玩家角色
	roleMgr := manager.GetPlayerRoleManager()
	playerRole := roleMgr.GetBySession(sessionId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found")
	}

	// 获取NPC配置
	configMgr := jsonconf.GetConfigManager()
	npcConfig := configMgr.GetNPCSceneConfig(req.NpcId)
	if npcConfig == nil {
		resp := &protocol.S2CTalkToNPCResultReq{
			Success: false,
			Message: "NPC不存在",
			NpcId:   req.NpcId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CTalkToNPCResult), resp)
	}

	// 触发任务事件（和NPC对话）
	questSys := GetQuestSys(ctx)
	if questSys != nil {
		questSys.UpdateQuestProgressByType(ctx, uint32(protocol.QuestType_QuestTypeTalkToNPC), req.NpcId, 1)
	}

	// 发送对话结果
	resp := &protocol.S2CTalkToNPCResultReq{
		Success: true,
		Message: "对话成功",
		NpcId:   req.NpcId,
	}

	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CTalkToNPCResult), resp)
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysQuest), func() iface.ISystem {
		return NewQuestSys()
	})
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2STalkToNPC), handleTalkToNPC)
	})
}
