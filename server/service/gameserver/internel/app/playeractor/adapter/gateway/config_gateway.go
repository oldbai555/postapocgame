package gateway

import (
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
)

// ConfigGatewayImpl 配置访问实现
type ConfigGatewayImpl struct{}

// NewConfigGateway 创建配置 Gateway
func NewConfigGateway() interfaces.ConfigManager {
	return &ConfigGatewayImpl{}
}

// GetBagConfig 获取背包配置
func (g *ConfigGatewayImpl) GetBagConfig(bagType uint32) *jsonconf.BagConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetBagConfig(bagType)
}

// GetItemConfig 获取物品配置
func (g *ConfigGatewayImpl) GetItemConfig(itemId uint32) *jsonconf.ItemConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetItemConfig(itemId)
}

// GetSkillConfig 获取技能配置
func (g *ConfigGatewayImpl) GetSkillConfig(skillId uint32) *jsonconf.SkillConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetSkillConfig(skillId)
}

// GetLevelConfig 获取等级配置
func (g *ConfigGatewayImpl) GetLevelConfig(level uint32) *jsonconf.LevelConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetLevelConfig(level)
}

// GetQuestConfig 获取任务配置
func (g *ConfigGatewayImpl) GetQuestConfig(questId uint32) *jsonconf.QuestConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetQuestConfig(questId)
}

// GetDungeonConfig 获取副本配置
func (g *ConfigGatewayImpl) GetDungeonConfig(dungeonId uint32) *jsonconf.DungeonConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetDungeonConfig(dungeonId)
}

// GetItemUseEffectConfig 获取物品使用效果配置
func (g *ConfigGatewayImpl) GetItemUseEffectConfig(itemId uint32) *jsonconf.ItemUseEffectConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetItemUseEffectConfig(itemId)
}

// GetJobConfig 获取职业配置
func (g *ConfigGatewayImpl) GetJobConfig(jobId uint32) *jsonconf.JobConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetJobConfig(jobId)
}

// GetEquipSetConfig 获取装备套装配置
func (g *ConfigGatewayImpl) GetEquipSetConfig(setId uint32) *jsonconf.EquipSetConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetEquipSetConfig(setId)
}

// GetQuestConfigsByType 根据任务类型获取任务配置列表
func (g *ConfigGatewayImpl) GetQuestConfigsByType(questType uint32) []*jsonconf.QuestConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetQuestConfigsByType(questType)
}

// GetNPCSceneConfig 获取NPC场景配置
func (g *ConfigGatewayImpl) GetNPCSceneConfig(npcId uint32) *jsonconf.NPCSceneConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetNPCSceneConfig(npcId)
}

// GetShopConfig 获取商城配置
func (g *ConfigGatewayImpl) GetShopConfig(itemId uint32) *jsonconf.ShopConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetShopConfig(itemId)
}

// GetConsumeConfig 获取消耗配置
func (g *ConfigGatewayImpl) GetConsumeConfig(consumeId uint32) *jsonconf.ConsumeConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetConsumeConfig(consumeId)
}

// GetRewardConfig 获取奖励配置
func (g *ConfigGatewayImpl) GetRewardConfig(rewardId uint32) *jsonconf.RewardConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetRewardConfig(rewardId)
}

// GetItemRecycleConfig 获取物品回收配置
func (g *ConfigGatewayImpl) GetItemRecycleConfig(itemId uint32) *jsonconf.ItemRecycleConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetItemRecycleConfig(itemId)
}

// GetMailTemplateConfig 获取邮件模板配置
func (g *ConfigGatewayImpl) GetMailTemplateConfig(templateId uint32) *jsonconf.MailTemplateConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetMailTemplateConfig(templateId)
}

// GetSensitiveWordConfig 获取敏感词配置
func (g *ConfigGatewayImpl) GetSensitiveWordConfig() []string {
	configMgr := jsonconf.GetConfigManager()
	cfg := configMgr.GetSensitiveWordConfig()
	if cfg == nil || len(cfg.Words) == 0 {
		return nil
	}
	words := make([]string, len(cfg.Words))
	copy(words, cfg.Words)
	return words
}

// GetVipConfig 获取VIP配置
func (g *ConfigGatewayImpl) GetVipConfig(level uint32) *jsonconf.VipConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetVipConfig(level)
}

// GetDailyActivityRewardConfig 获取日常活跃奖励配置
func (g *ConfigGatewayImpl) GetDailyActivityRewardConfig(rewardId uint32) *jsonconf.DailyActivityRewardConfig {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetDailyActivityRewardConfig(rewardId)
}
