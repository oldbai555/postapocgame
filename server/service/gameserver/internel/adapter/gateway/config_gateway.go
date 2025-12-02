package gateway

import (
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// ConfigGatewayImpl 配置访问实现
type ConfigGatewayImpl struct{}

// NewConfigGateway 创建配置 Gateway
func NewConfigGateway() interfaces.ConfigManager {
	return &ConfigGatewayImpl{}
}

// GetBagConfig 获取背包配置
func (g *ConfigGatewayImpl) GetBagConfig(bagType uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetBagConfig(bagType)
	return config, ok
}

// GetItemConfig 获取物品配置
func (g *ConfigGatewayImpl) GetItemConfig(itemId uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetItemConfig(itemId)
	return config, ok
}

// GetSkillConfig 获取技能配置
func (g *ConfigGatewayImpl) GetSkillConfig(skillId uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetSkillConfig(skillId)
	return config, ok
}

// GetLevelConfig 获取等级配置
func (g *ConfigGatewayImpl) GetLevelConfig(level uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetLevelConfig(level)
	return config, ok
}

// GetQuestConfig 获取任务配置
func (g *ConfigGatewayImpl) GetQuestConfig(questId uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetQuestConfig(questId)
	return config, ok
}

// GetDungeonConfig 获取副本配置
func (g *ConfigGatewayImpl) GetDungeonConfig(dungeonId uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetDungeonConfig(dungeonId)
	return config, ok
}

// GetItemUseEffectConfig 获取物品使用效果配置
func (g *ConfigGatewayImpl) GetItemUseEffectConfig(itemId uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetItemUseEffectConfig(itemId)
	return config, ok
}

// GetJobConfig 获取职业配置
func (g *ConfigGatewayImpl) GetJobConfig(jobId uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetJobConfig(jobId)
	return config, ok
}

// GetEquipSetConfig 获取装备套装配置
func (g *ConfigGatewayImpl) GetEquipSetConfig(setId uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetEquipSetConfig(setId)
	return config, ok
}

// GetQuestConfigsByType 根据任务类型获取任务配置列表
func (g *ConfigGatewayImpl) GetQuestConfigsByType(questType uint32) []interface{} {
	configMgr := jsonconf.GetConfigManager()
	configs := configMgr.GetQuestConfigsByType(questType)
	result := make([]interface{}, len(configs))
	for i, cfg := range configs {
		result[i] = cfg
	}
	return result
}

// GetNPCSceneConfig 获取NPC场景配置
func (g *ConfigGatewayImpl) GetNPCSceneConfig(npcId uint32) interface{} {
	configMgr := jsonconf.GetConfigManager()
	return configMgr.GetNPCSceneConfig(npcId)
}

// GetShopConfig 获取商城配置
func (g *ConfigGatewayImpl) GetShopConfig(itemId uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetShopConfig(itemId)
	return config, ok
}

// GetConsumeConfig 获取消耗配置
func (g *ConfigGatewayImpl) GetConsumeConfig(consumeId uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetConsumeConfig(consumeId)
	return config, ok
}

// GetRewardConfig 获取奖励配置
func (g *ConfigGatewayImpl) GetRewardConfig(rewardId uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetRewardConfig(rewardId)
	return config, ok
}

// GetItemRecycleConfig 获取物品回收配置
func (g *ConfigGatewayImpl) GetItemRecycleConfig(itemId uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetItemRecycleConfig(itemId)
	return config, ok
}

// GetMailTemplateConfig 获取邮件模板配置
func (g *ConfigGatewayImpl) GetMailTemplateConfig(templateId uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetMailTemplateConfig(templateId)
	return config, ok
}

// GetSensitiveWordConfig 获取敏感词配置
func (g *ConfigGatewayImpl) GetSensitiveWordConfig() ([]string, bool) {
	configMgr := jsonconf.GetConfigManager()
	cfg := configMgr.GetSensitiveWordConfig()
	if cfg == nil || len(cfg.Words) == 0 {
		return nil, false
	}
	words := make([]string, len(cfg.Words))
	copy(words, cfg.Words)
	return words, true
}

// GetVipConfig 获取VIP配置
func (g *ConfigGatewayImpl) GetVipConfig(level uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetVipConfig(level)
	return config, ok
}

// GetDailyActivityRewardConfig 获取日常活跃奖励配置
func (g *ConfigGatewayImpl) GetDailyActivityRewardConfig(rewardId uint32) (interface{}, bool) {
	configMgr := jsonconf.GetConfigManager()
	config, ok := configMgr.GetDailyActivityRewardConfig(rewardId)
	return config, ok
}
