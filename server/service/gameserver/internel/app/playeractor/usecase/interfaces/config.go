package interfaces

import "postapocgame/server/internal/jsonconf"

// ConfigManager 配置管理器接口（Use Case 层定义）
type ConfigManager interface {
	// GetBagConfig 获取背包配置
	GetBagConfig(bagType uint32) *jsonconf.BagConfig

	// GetItemConfig 获取物品配置
	GetItemConfig(itemId uint32) *jsonconf.ItemConfig

	// GetSkillConfig 获取技能配置
	GetSkillConfig(skillId uint32) *jsonconf.SkillConfig

	// GetLevelConfig 获取等级配置
	GetLevelConfig(level uint32) *jsonconf.LevelConfig

	// GetQuestConfig 获取任务配置
	GetQuestConfig(questId uint32) *jsonconf.QuestConfig

	// GetDungeonConfig 获取副本配置
	GetDungeonConfig(dungeonId uint32) *jsonconf.DungeonConfig

	// GetItemUseEffectConfig 获取物品使用效果配置
	GetItemUseEffectConfig(itemId uint32) *jsonconf.ItemUseEffectConfig

	// GetJobConfig 获取职业配置
	GetJobConfig(jobId uint32) *jsonconf.JobConfig

	// GetEquipSetConfig 获取装备套装配置
	GetEquipSetConfig(setId uint32) *jsonconf.EquipSetConfig

	// GetQuestConfigsByType 根据任务类型获取任务配置列表
	GetQuestConfigsByType(questType uint32) []*jsonconf.QuestConfig

	// GetNPCSceneConfig 获取NPC场景配置
	GetNPCSceneConfig(npcId uint32) *jsonconf.NPCSceneConfig

	// GetShopConfig 获取商城配置
	GetShopConfig(itemId uint32) *jsonconf.ShopConfig

	// GetConsumeConfig 获取消耗配置
	GetConsumeConfig(consumeId uint32) *jsonconf.ConsumeConfig

	// GetRewardConfig 获取奖励配置
	GetRewardConfig(rewardId uint32) *jsonconf.RewardConfig

	// GetItemRecycleConfig 获取物品回收配置
	GetItemRecycleConfig(itemId uint32) *jsonconf.ItemRecycleConfig

	// GetSensitiveWordConfig 获取敏感词配置
	GetSensitiveWordConfig() []string

	// GetMailTemplateConfig 获取邮件模板配置
	GetMailTemplateConfig(templateId uint32) *jsonconf.MailTemplateConfig

	// GetVipConfig 获取VIP配置
	GetVipConfig(level uint32) *jsonconf.VipConfig

	// GetDailyActivityRewardConfig 获取日常活跃奖励配置
	GetDailyActivityRewardConfig(rewardId uint32) *jsonconf.DailyActivityRewardConfig
}
