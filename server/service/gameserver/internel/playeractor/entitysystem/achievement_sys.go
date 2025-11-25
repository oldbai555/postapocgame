package entitysystem

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/iface"
)

// AchievementSys 成就系统
type AchievementSys struct {
	*BaseSystem
	achievementData *protocol.SiAchievementData
}

// NewAchievementSys 创建成就系统
func NewAchievementSys() *AchievementSys {
	return &AchievementSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysAchievement)),
	}
}

func GetAchievementSys(ctx context.Context) *AchievementSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysAchievement))
	if system == nil {
		return nil
	}
	achievementSys, ok := system.(*AchievementSys)
	if !ok || !achievementSys.IsOpened() {
		return nil
	}
	return achievementSys
}

// OnInit 系统初始化
func (as *AchievementSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("achievement sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果achievement_data不存在，则初始化
	if binaryData.AchievementData == nil {
		binaryData.AchievementData = &protocol.SiAchievementData{
			Achievements: make([]*protocol.AchievementData, 0),
		}
	}
	as.achievementData = binaryData.AchievementData

	// 如果Achievements为空，初始化为空切片
	if as.achievementData.Achievements == nil {
		as.achievementData.Achievements = make([]*protocol.AchievementData, 0)
	}

	log.Infof("AchievementSys initialized: AchievementCount=%d", len(as.achievementData.Achievements))
}

// GetAchievement 获取指定成就
func (as *AchievementSys) GetAchievement(achievementId uint32) *protocol.AchievementData {
	if as.achievementData == nil || as.achievementData.Achievements == nil {
		return nil
	}
	for _, achievement := range as.achievementData.Achievements {
		if achievement != nil && achievement.AchievementId == achievementId {
			return achievement
		}
	}
	return nil
}

// GetOrCreateAchievement 获取或创建成就数据
func (as *AchievementSys) GetOrCreateAchievement(achievementId uint32) *protocol.AchievementData {
	achievement := as.GetAchievement(achievementId)
	if achievement != nil {
		return achievement
	}

	// 创建新的成就数据
	achievement = &protocol.AchievementData{
		AchievementId: achievementId,
		Progress:      make([]uint32, 0),
		Completed:     false,
		CompleteTime:  0,
	}

	// 获取成就配置，初始化进度数组
	achievementConfig, ok := jsonconf.GetConfigManager().GetAchievementConfig(achievementId)
	if ok {
		achievement.Progress = make([]uint32, len(achievementConfig.Targets))
		for i := range achievement.Progress {
			achievement.Progress[i] = 0
		}
	}

	// 添加到成就列表
	if as.achievementData.Achievements == nil {
		as.achievementData.Achievements = make([]*protocol.AchievementData, 0)
	}
	as.achievementData.Achievements = append(as.achievementData.Achievements, achievement)

	return achievement
}

// UpdateAchievementProgress 更新成就进度（按目标索引）
func (as *AchievementSys) UpdateAchievementProgress(ctx context.Context, achievementId uint32, targetIndex uint32, progress uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	// 获取或创建成就数据
	achievement := as.GetOrCreateAchievement(achievementId)

	// 检查成就配置
	achievementConfig, ok := jsonconf.GetConfigManager().GetAchievementConfig(achievementId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "achievement config not found: %d", achievementId)
	}

	// 检查目标索引
	if int(targetIndex) >= len(achievementConfig.Targets) {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "invalid target index: %d", targetIndex)
	}

	// 如果已完成，不再更新
	if achievement.Completed {
		return nil
	}

	// 更新进度（不能超过目标数量）
	target := achievementConfig.Targets[targetIndex]
	if progress > target.Count {
		progress = target.Count
	}

	// 确保Progress数组足够大
	for int(targetIndex) >= len(achievement.Progress) {
		achievement.Progress = append(achievement.Progress, 0)
	}

	achievement.Progress[targetIndex] = progress

	log.Infof("Achievement progress updated: RoleID=%d, AchievementID=%d, TargetIndex=%d, Progress=%d/%d",
		playerRole.GetPlayerRoleId(), achievementId, targetIndex, progress, target.Count)

	// 检查成就是否完成
	if as.isAchievementCompleted(achievementId) {
		as.completeAchievement(ctx, achievementId)
	}

	return nil
}

// UpdateAchievementProgressByType 根据成就类型更新进度（自动匹配符合条件的成就目标）
func (as *AchievementSys) UpdateAchievementProgressByType(ctx context.Context, achievementType uint32, targetId uint32, count uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	// 获取所有成就配置
	configManager := jsonconf.GetConfigManager()
	allAchievements := configManager.GetAllAchievementConfigs()

	// 遍历所有成就配置
	for _, achievementConfig := range allAchievements {
		if achievementConfig == nil {
			continue
		}

		// 检查成就类型是否匹配
		if achievementConfig.Type != achievementType {
			continue
		}

		// 获取或创建成就数据
		achievement := as.GetOrCreateAchievement(achievementConfig.AchievementId)

		// 如果已完成，跳过
		if achievement.Completed {
			continue
		}

		// 遍历成就的所有目标
		for targetIndex, target := range achievementConfig.Targets {
			// 根据成就类型进行匹配检查
			matched := false
			switch achievementType {
			case 1: // 等级成就
				matched = true
			case 2: // 任务成就
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
			case 3: // 副本成就
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
			case 4: // 战斗成就
				matched = true
			}

			if matched {
				// 确保Progress数组足够大
				for int(targetIndex) >= len(achievement.Progress) {
					achievement.Progress = append(achievement.Progress, 0)
				}

				// 增加进度（不能超过目标数量）
				newProgress := achievement.Progress[targetIndex] + count
				if newProgress > target.Count {
					newProgress = target.Count
				}
				achievement.Progress[targetIndex] = newProgress

				log.Infof("Achievement progress updated by type: RoleID=%d, AchievementID=%d, TargetIndex=%d, Progress=%d/%d, Type=%d, TargetId=%d",
					playerRole.GetPlayerRoleId(), achievement.AchievementId, targetIndex, newProgress, target.Count, achievementType, targetId)

				// 检查成就是否完成
				if as.isAchievementCompleted(achievement.AchievementId) {
					as.completeAchievement(ctx, achievement.AchievementId)
				}
			}
		}
	}

	return nil
}

// isAchievementCompleted 检查成就是否完成
func (as *AchievementSys) isAchievementCompleted(achievementId uint32) bool {
	achievement := as.GetAchievement(achievementId)
	if achievement == nil {
		return false
	}

	if achievement.Completed {
		return true
	}

	achievementConfig, ok := jsonconf.GetConfigManager().GetAchievementConfig(achievementId)
	if !ok {
		return false
	}

	// 检查所有目标是否完成
	for i, target := range achievementConfig.Targets {
		if i >= len(achievement.Progress) {
			return false
		}
		if achievement.Progress[i] < target.Count {
			return false
		}
	}

	return true
}

// completeAchievement 完成成就
func (as *AchievementSys) completeAchievement(ctx context.Context, achievementId uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	achievement := as.GetAchievement(achievementId)
	if achievement == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "achievement not found: %d", achievementId)
	}

	// 如果已完成，不再处理
	if achievement.Completed {
		return nil
	}

	// 获取成就配置
	achievementConfig, ok := jsonconf.GetConfigManager().GetAchievementConfig(achievementId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "achievement config not found: %d", achievementId)
	}

	// 标记为已完成
	achievement.Completed = true
	achievement.CompleteTime = servertime.Now().Unix()

	// 发放经验奖励
	if achievementConfig.ExpReward > 0 {
		levelSys := GetLevelSys(ctx)
		if levelSys != nil {
			if err := levelSys.AddExp(ctx, achievementConfig.ExpReward); err != nil {
				log.Errorf("AddExp failed: %v", err)
			}
		}
	}

	// 发放物品奖励
	if len(achievementConfig.Rewards) > 0 {
		rewards := make([]*jsonconf.ItemAmount, 0, len(achievementConfig.Rewards))
		for _, reward := range achievementConfig.Rewards {
			rewards = append(rewards, &jsonconf.ItemAmount{
				ItemType: uint32(reward.Type),
				ItemId:   reward.ItemId,
				Count:    int64(reward.Count),
				Bind:     1, // 成就奖励默认绑定
			})
		}
		if err := playerRole.GrantRewards(ctx, rewards); err != nil {
			log.Errorf("GrantRewards failed: %v", err)
			return customerr.Wrap(err)
		}
	}

	log.Infof("Achievement completed: RoleID=%d, AchievementID=%d", playerRole.GetPlayerRoleId(), achievementId)
	return nil
}

// GetAchievementData 获取成就数据
func (as *AchievementSys) GetAchievementData() *protocol.SiAchievementData {
	return as.achievementData
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysAchievement), func() iface.ISystem {
		return NewAchievementSys()
	})
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
	})
}
