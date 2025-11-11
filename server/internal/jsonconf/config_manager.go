package jsonconf

import (
	"fmt"
	"os"
	"path/filepath"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"sync"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	configPath string
	mu         sync.RWMutex

	// 各类配置
	itemConfigs         map[uint32]*ItemConfig
	levelConfigs        map[uint32]*LevelConfig
	vipConfigs          map[uint32]*VipConfig
	moneyConfigs        map[uint32]*MoneyConfig
	questConfigs        map[uint32]*QuestConfig
	mailTemplateConfigs map[uint32]*MailTemplateConfig
	monsterConfigs      map[uint32]*MonsterConfig
	skillConfigs        map[uint32]*SkillConfig
	buffConfigs         map[uint32]*BuffConfig
}

var (
	globalConfigManager *ConfigManager
	configOnce          sync.Once
)

// GetConfigManager 获取全局配置管理器
func GetConfigManager() *ConfigManager {
	configOnce.Do(func() {
		globalConfigManager = &ConfigManager{
			itemConfigs:         make(map[uint32]*ItemConfig),
			levelConfigs:        make(map[uint32]*LevelConfig),
			vipConfigs:          make(map[uint32]*VipConfig),
			moneyConfigs:        make(map[uint32]*MoneyConfig),
			questConfigs:        make(map[uint32]*QuestConfig),
			mailTemplateConfigs: make(map[uint32]*MailTemplateConfig),
			monsterConfigs:      make(map[uint32]*MonsterConfig),
			skillConfigs:        make(map[uint32]*SkillConfig),
			buffConfigs:         make(map[uint32]*BuffConfig),
		}
	})
	return globalConfigManager
}

// Init 初始化配置管理器
func (cm *ConfigManager) Init(configPath string) error {
	cm.configPath = configPath

	// 加载所有配置
	if err := cm.LoadAllConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	log.Infof("ConfigManager initialized, configPath=%s", configPath)
	return nil
}

// LoadAllConfigs 加载所有配置
func (cm *ConfigManager) LoadAllConfigs() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 加载道具配置
	if err := cm.loadItemConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载等级配置
	if err := cm.loadLevelConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载VIP配置
	if err := cm.loadVipConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载货币配置
	if err := cm.loadMoneyConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载任务配置
	if err := cm.loadQuestConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载邮件模板配置
	if err := cm.loadMailTemplateConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载怪物配置
	if err := cm.loadMonsterConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载技能配置
	if err := cm.loadSkillConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载Buff配置
	if err := cm.loadBuffConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	log.Infof("All configs loaded successfully")
	return nil
}

// loadItemConfigs 加载道具配置
func (cm *ConfigManager) loadItemConfigs() error {
	filePath := filepath.Join(cm.configPath, "item_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read item config failed: %w", err)
	}

	var configs []*ItemConfig
	if err := tool.JsonUnmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal item config failed: %w", err)
	}

	cm.itemConfigs = make(map[uint32]*ItemConfig)
	for _, cfg := range configs {
		cm.itemConfigs[cfg.ItemId] = cfg
	}

	log.Infof("Loaded %d item configs", len(cm.itemConfigs))
	return nil
}

// loadLevelConfigs 加载等级配置
func (cm *ConfigManager) loadLevelConfigs() error {
	filePath := filepath.Join(cm.configPath, "level_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read level config failed: %w", err)
	}

	var configs []*LevelConfig
	if err := tool.JsonUnmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal level config failed: %w", err)
	}

	cm.levelConfigs = make(map[uint32]*LevelConfig)
	for _, cfg := range configs {
		cm.levelConfigs[cfg.Level] = cfg
	}

	log.Infof("Loaded %d level configs", len(cm.levelConfigs))
	return nil
}

// loadVipConfigs 加载VIP配置
func (cm *ConfigManager) loadVipConfigs() error {
	filePath := filepath.Join(cm.configPath, "vip_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read vip config failed: %w", err)
	}

	var configs []*VipConfig
	if err := tool.JsonUnmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal vip config failed: %w", err)
	}

	cm.vipConfigs = make(map[uint32]*VipConfig)
	for _, cfg := range configs {
		cm.vipConfigs[cfg.Level] = cfg
	}

	log.Infof("Loaded %d vip configs", len(cm.vipConfigs))
	return nil
}

// loadMoneyConfigs 加载货币配置
func (cm *ConfigManager) loadMoneyConfigs() error {
	filePath := filepath.Join(cm.configPath, "money_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read money config failed: %w", err)
	}

	var configs []*MoneyConfig
	if err := tool.JsonUnmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal money config failed: %w", err)
	}

	cm.moneyConfigs = make(map[uint32]*MoneyConfig)
	for _, cfg := range configs {
		cm.moneyConfigs[cfg.MoneyId] = cfg
	}

	log.Infof("Loaded %d money configs", len(cm.moneyConfigs))
	return nil
}

// loadQuestConfigs 加载任务配置
func (cm *ConfigManager) loadQuestConfigs() error {
	filePath := filepath.Join(cm.configPath, "quest_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read quest config failed: %w", err)
	}

	var configs []*QuestConfig
	if err := tool.JsonUnmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal quest config failed: %w", err)
	}

	cm.questConfigs = make(map[uint32]*QuestConfig)
	for _, cfg := range configs {
		cm.questConfigs[cfg.QuestId] = cfg
	}

	log.Infof("Loaded %d quest configs", len(cm.questConfigs))
	return nil
}

// loadMailTemplateConfigs 加载邮件模板配置
func (cm *ConfigManager) loadMailTemplateConfigs() error {
	filePath := filepath.Join(cm.configPath, "mail_template_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read mail template config failed: %w", err)
	}

	var configs []*MailTemplateConfig
	if err := tool.JsonUnmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal mail template config failed: %w", err)
	}

	cm.mailTemplateConfigs = make(map[uint32]*MailTemplateConfig)
	for _, cfg := range configs {
		cm.mailTemplateConfigs[cfg.TemplateId] = cfg
	}

	log.Infof("Loaded %d mail template configs", len(cm.mailTemplateConfigs))
	return nil
}

// loadMonsterConfigs 加载怪物配置
func (cm *ConfigManager) loadMonsterConfigs() error {
	filePath := filepath.Join(cm.configPath, "monster_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read monster config failed: %w", err)
	}

	var configs []*MonsterConfig
	if err := tool.JsonUnmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal monster config failed: %w", err)
	}

	cm.monsterConfigs = make(map[uint32]*MonsterConfig)
	for _, cfg := range configs {
		cm.monsterConfigs[cfg.MonsterId] = cfg
	}

	log.Infof("Loaded %d monster configs", len(cm.monsterConfigs))
	return nil
}

// loadSkillConfigs 加载技能配置
func (cm *ConfigManager) loadSkillConfigs() error {
	filePath := filepath.Join(cm.configPath, "skill_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read skill config failed: %w", err)
	}

	var configs []*SkillConfig
	if err := tool.JsonUnmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal skill config failed: %w", err)
	}

	cm.skillConfigs = make(map[uint32]*SkillConfig)
	for _, cfg := range configs {
		cm.skillConfigs[cfg.SkillId] = cfg
	}

	log.Infof("Loaded %d skill configs", len(cm.skillConfigs))
	return nil
}

// loadBuffConfigs 加载Buff配置
func (cm *ConfigManager) loadBuffConfigs() error {
	filePath := filepath.Join(cm.configPath, "buff_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read buff config failed: %w", err)
	}

	var configs []*BuffConfig
	if err := tool.JsonUnmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal buff config failed: %w", err)
	}

	cm.buffConfigs = make(map[uint32]*BuffConfig)
	for _, cfg := range configs {
		cm.buffConfigs[cfg.BuffId] = cfg
	}

	log.Infof("Loaded %d buff configs", len(cm.buffConfigs))
	return nil
}

// GetItemConfig 获取道具配置
func (cm *ConfigManager) GetItemConfig(itemId uint32) (*ItemConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.itemConfigs[itemId]
	return cfg, ok
}

// GetLevelConfig 获取等级配置
func (cm *ConfigManager) GetLevelConfig(level uint32) (*LevelConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.levelConfigs[level]
	return cfg, ok
}

// GetVipConfig 获取VIP配置
func (cm *ConfigManager) GetVipConfig(level uint32) (*VipConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.vipConfigs[level]
	return cfg, ok
}

// GetMoneyConfig 获取货币配置
func (cm *ConfigManager) GetMoneyConfig(moneyId uint32) (*MoneyConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.moneyConfigs[moneyId]
	return cfg, ok
}

// GetQuestConfig 获取任务配置
func (cm *ConfigManager) GetQuestConfig(questId uint32) (*QuestConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.questConfigs[questId]
	return cfg, ok
}

// GetMailTemplateConfig 获取邮件模板配置
func (cm *ConfigManager) GetMailTemplateConfig(templateId uint32) (*MailTemplateConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.mailTemplateConfigs[templateId]
	return cfg, ok
}

// Reload 热加载配置
func (cm *ConfigManager) Reload() error {
	log.Infof("Reloading all configs...")
	return cm.LoadAllConfigs()
}

// GetMonsterConfig 获取怪物配置
func (cm *ConfigManager) GetMonsterConfig(monsterId uint32) (*MonsterConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.monsterConfigs[monsterId]
	return cfg, ok
}

// GetSkillConfig 获取技能配置
func (cm *ConfigManager) GetSkillConfig(skillId uint32) (*SkillConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.skillConfigs[skillId]
	return cfg, ok
}

// GetBuffConfig 获取Buff配置
func (cm *ConfigManager) GetBuffConfig(buffId uint32) (*BuffConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.buffConfigs[buffId]
	return cfg, ok
}
