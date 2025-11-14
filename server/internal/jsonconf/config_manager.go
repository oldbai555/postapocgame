package jsonconf

import (
	"fmt"
	"os"
	"path/filepath"
	"postapocgame/server/internal"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"sync"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	configPath string
	mu         sync.RWMutex

	// 各类配置
	itemConfigs                map[uint32]*ItemConfig
	levelConfigs               map[uint32]*LevelConfig
	vipConfigs                 map[uint32]*VipConfig
	moneyConfigs               map[uint32]*MoneyConfig
	questConfigs               map[uint32]*QuestConfig
	mailTemplateConfigs        map[uint32]*MailTemplateConfig
	monsterConfigs             map[uint32]*MonsterConfig
	skillConfigs               map[uint32]*SkillConfig
	buffConfigs                map[uint32]*BuffConfig
	dungeonConfigs             map[uint32]*DungeonConfig
	consumeConfigs             map[uint32]*ConsumeConfig
	rewardConfigs              map[uint32]*RewardConfig
	equipUpgradeConfigs        map[uint32]*EquipUpgradeConfig
	itemUseEffectConfigs       map[uint32]*ItemUseEffectConfig
	itemRecycleConfigs         map[uint32]*ItemRecycleConfig
	shopConfigs                map[uint32]*ShopConfig // key: itemId
	jobConfigs                 map[uint32]*JobConfig
	sceneConfigs               map[uint32]*SceneConfig
	monsterSceneConfigs        []*MonsterSceneConfig // 列表形式，因为一个场景可能有多个怪物配置
	npcSceneConfigs            []*NPCSceneConfig     // 列表形式，因为一个场景可能有多个NPC
	teleportConfigs            map[uint32]*TeleportConfig
	dailyActivityRewardConfigs map[uint32]*DailyActivityRewardConfig
	achievementConfigs         map[uint32]*AchievementConfig
	equipRefineConfigs         map[uint32]*EquipRefineConfig
	equipEnchantConfigs        map[uint32]*EquipEnchantConfig
	equipSetConfigs            map[uint32]*EquipSetConfig
	bagConfigs                 map[uint32]*BagConfig // key: bagType
}

var (
	globalConfigManager *ConfigManager
	configOnce          sync.Once
)

// GetConfigManager 获取全局配置管理器
func GetConfigManager() *ConfigManager {
	configOnce.Do(func() {
		globalConfigManager = &ConfigManager{
			itemConfigs:                make(map[uint32]*ItemConfig),
			levelConfigs:               make(map[uint32]*LevelConfig),
			vipConfigs:                 make(map[uint32]*VipConfig),
			moneyConfigs:               make(map[uint32]*MoneyConfig),
			questConfigs:               make(map[uint32]*QuestConfig),
			mailTemplateConfigs:        make(map[uint32]*MailTemplateConfig),
			monsterConfigs:             make(map[uint32]*MonsterConfig),
			skillConfigs:               make(map[uint32]*SkillConfig),
			buffConfigs:                make(map[uint32]*BuffConfig),
			dungeonConfigs:             make(map[uint32]*DungeonConfig),
			consumeConfigs:             make(map[uint32]*ConsumeConfig),
			rewardConfigs:              make(map[uint32]*RewardConfig),
			equipUpgradeConfigs:        make(map[uint32]*EquipUpgradeConfig),
			itemUseEffectConfigs:       make(map[uint32]*ItemUseEffectConfig),
			itemRecycleConfigs:         make(map[uint32]*ItemRecycleConfig),
			shopConfigs:                make(map[uint32]*ShopConfig),
			jobConfigs:                 make(map[uint32]*JobConfig),
			sceneConfigs:               make(map[uint32]*SceneConfig),
			monsterSceneConfigs:        make([]*MonsterSceneConfig, 0),
			npcSceneConfigs:            make([]*NPCSceneConfig, 0),
			teleportConfigs:            make(map[uint32]*TeleportConfig),
			dailyActivityRewardConfigs: make(map[uint32]*DailyActivityRewardConfig),
			achievementConfigs:         make(map[uint32]*AchievementConfig),
			equipRefineConfigs:         make(map[uint32]*EquipRefineConfig),
			equipEnchantConfigs:        make(map[uint32]*EquipEnchantConfig),
			equipSetConfigs:            make(map[uint32]*EquipSetConfig),
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

	// 加载副本配置
	if err := cm.loadDungeonConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载通用消耗配置
	if err := cm.loadConsumeConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载通用奖励配置
	if err := cm.loadRewardConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载装备强化配置
	if err := cm.loadEquipUpgradeConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载物品使用效果配置
	if err := cm.loadItemUseEffectConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载物品回收配置
	if err := cm.loadItemRecycleConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载商城配置
	if err := cm.loadShopConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载职业配置
	if err := cm.loadJobConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载场景配置
	if err := cm.loadSceneConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载怪物场景配置
	if err := cm.loadMonsterSceneConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载NPC场景配置
	if err := cm.loadNPCSceneConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载传送点配置
	if err := cm.loadTeleportConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载日常活跃奖励配置
	if err := cm.loadDailyActivityRewardConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载成就配置
	if err := cm.loadAchievementConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	// 加载装备精炼配置（可选）
	if err := cm.loadEquipRefineConfigs(); err != nil {
		log.Warnf("load equip refine configs failed: %v", err)
		// 不返回错误，允许配置不存在
	}

	// 加载装备附魔配置（可选）
	if err := cm.loadEquipEnchantConfigs(); err != nil {
		log.Warnf("load equip enchant configs failed: %v", err)
		// 不返回错误，允许配置不存在
	}

	// 加载装备套装配置（可选）
	if err := cm.loadEquipSetConfigs(); err != nil {
		log.Warnf("load equip set configs failed: %v", err)
		// 不返回错误，允许配置不存在
	}

	// 加载背包配置（可选）
	if err := cm.loadBagConfigs(); err != nil {
		log.Warnf("load bag configs failed: %v", err)
		// 不返回错误，允许配置不存在
	}

	log.Infof("All configs loaded successfully")
	return nil
}

// loadItemConfigs 加载道具配置
func (cm *ConfigManager) loadItemConfigs() error {
	filePath := filepath.Join(cm.configPath, "itemconfig.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read item config failed: %w", err)
	}

	var configs []*ItemConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
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
	filePath := filepath.Join(cm.configPath, "levelconfig.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read level config failed: %w", err)
	}

	var configs []*LevelConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
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
	filePath := filepath.Join(cm.configPath, "vipconfig.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read vip config failed: %w", err)
	}

	var configs []*VipConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
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
	filePath := filepath.Join(cm.configPath, "moneyconfig.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read money config failed: %w", err)
	}

	var configs []*MoneyConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
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
	filePath := filepath.Join(cm.configPath, "questconfig.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read quest config failed: %w", err)
	}

	var configs []*QuestConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
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
	filePath := filepath.Join(cm.configPath, "mailtemplateconfig.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read mail template config failed: %w", err)
	}

	var configs []*MailTemplateConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
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
	filePath := filepath.Join(cm.configPath, "monsterconfig.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read monster config failed: %w", err)
	}

	var configs []*MonsterConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
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
	filePath := filepath.Join(cm.configPath, "skillconfig.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read skill config failed: %w", err)
	}

	var configs []*SkillConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
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
	filePath := filepath.Join(cm.configPath, "buffconfig.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read buff config failed: %w", err)
	}

	var configs []*BuffConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal buff config failed: %w", err)
	}

	cm.buffConfigs = make(map[uint32]*BuffConfig)
	for _, cfg := range configs {
		cm.buffConfigs[cfg.BuffId] = cfg
	}

	log.Infof("Loaded %d buff configs", len(cm.buffConfigs))
	return nil
}

func (cm *ConfigManager) loadConsumeConfigs() error {
	filePath := filepath.Join(cm.configPath, "consume_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warnf("consume config file not found: %s", filePath)
			return nil
		}
		return fmt.Errorf("read consume config failed: %w", err)
	}

	var configs []*ConsumeConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal consume config failed: %w", err)
	}

	cm.consumeConfigs = make(map[uint32]*ConsumeConfig)
	for _, cfg := range configs {
		cm.consumeConfigs[cfg.ConsumeId] = cfg
	}

	log.Infof("Loaded %d consume configs", len(cm.consumeConfigs))
	return nil
}

func (cm *ConfigManager) loadRewardConfigs() error {
	filePath := filepath.Join(cm.configPath, "reward_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warnf("reward config file not found: %s", filePath)
			return nil
		}
		return fmt.Errorf("read reward config failed: %w", err)
	}

	var configs []*RewardConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal reward config failed: %w", err)
	}

	cm.rewardConfigs = make(map[uint32]*RewardConfig)
	for _, cfg := range configs {
		cm.rewardConfigs[cfg.RewardId] = cfg
	}

	log.Infof("Loaded %d reward configs", len(cm.rewardConfigs))
	return nil
}

// loadEquipUpgradeConfigs 加载装备强化配置
func (cm *ConfigManager) loadEquipUpgradeConfigs() error {
	filePath := filepath.Join(cm.configPath, "equip_upgrade_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("equip_upgrade_config.json not found, using empty config")
			return nil
		}
		return fmt.Errorf("read equip upgrade config failed: %w", err)
	}

	var configs []*EquipUpgradeConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal equip upgrade config failed: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, cfg := range configs {
		if cfg != nil {
			cm.equipUpgradeConfigs[cfg.ItemId] = cfg
		}
	}

	log.Infof("Loaded %d equip upgrade configs", len(cm.equipUpgradeConfigs))
	return nil
}

// GetItemConfig 获取道具配置
func (cm *ConfigManager) GetItemConfig(itemId uint32) (*ItemConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.itemConfigs[itemId]
	return cfg, ok
}

// GetEquipUpgradeConfig 获取装备强化配置
func (cm *ConfigManager) GetEquipUpgradeConfig(itemId uint32) (*EquipUpgradeConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.equipUpgradeConfigs[itemId]
	return cfg, ok
}

// GetLevelConfig 获取等级配置
func (cm *ConfigManager) GetLevelConfig(level uint32) (*LevelConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.levelConfigs[level]
	return cfg, ok
}

// GetLevelAttrs 获取指定等级的属性（高等级覆盖低等级）
// 返回该等级及以下所有等级的属性合并结果，高等级属性会覆盖低等级同名属性
func (cm *ConfigManager) GetLevelAttrs(level uint32) map[uint32]uint64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[uint32]uint64)

	// 从1级到指定等级，逐级合并属性（高等级覆盖低等级）
	for l := uint32(1); l <= level; l++ {
		cfg, ok := cm.levelConfigs[l]
		if !ok {
			continue
		}

		// 合并该等级的属性（高等级会覆盖低等级）
		for _, attr := range cfg.Attrs {
			if attr != nil {
				result[attr.Type] = attr.Value
			}
		}
	}

	return result
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

// GetQuestConfigsByType 根据类型获取任务配置
func (cm *ConfigManager) GetQuestConfigsByType(questType uint32) []*QuestConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	result := make([]*QuestConfig, 0)
	for _, cfg := range cm.questConfigs {
		if cfg == nil {
			continue
		}
		if cfg.Type == questType {
			result = append(result, cfg)
		}
	}
	return result
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

// GetConsumeConfig 获取通用消耗配置
func (cm *ConfigManager) GetConsumeConfig(consumeId uint32) (*ConsumeConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.consumeConfigs[consumeId]
	return cfg, ok
}

// GetRewardConfig 获取通用奖励配置
func (cm *ConfigManager) GetRewardConfig(rewardId uint32) (*RewardConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.rewardConfigs[rewardId]
	return cfg, ok
}

// loadDungeonConfigs 加载副本配置
func (cm *ConfigManager) loadDungeonConfigs() error {
	filePath := filepath.Join(cm.configPath, "dungeon_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 副本配置可选，如果文件不存在也不报错
		log.Warnf("dungeon config file not found: %s", filePath)
		cm.dungeonConfigs = make(map[uint32]*DungeonConfig)
		return nil
	}

	var configs []*DungeonConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal dungeon config failed: %w", err)
	}

	cm.dungeonConfigs = make(map[uint32]*DungeonConfig)
	for _, cfg := range configs {
		cm.dungeonConfigs[cfg.DungeonID] = cfg
	}

	log.Infof("Loaded %d dungeon configs", len(cm.dungeonConfigs))
	return nil
}

// GetDungeonConfig 获取副本配置
func (cm *ConfigManager) GetDungeonConfig(dungeonId uint32) (*DungeonConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.dungeonConfigs[dungeonId]
	return cfg, ok
}

// loadItemUseEffectConfigs 加载物品使用效果配置
func (cm *ConfigManager) loadItemUseEffectConfigs() error {
	filePath := filepath.Join(cm.configPath, "item_use_effect_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("item_use_effect_config.json not found, using empty config")
			cm.itemUseEffectConfigs = make(map[uint32]*ItemUseEffectConfig)
			return nil
		}
		return fmt.Errorf("read item use effect config failed: %w", err)
	}

	var configs []*ItemUseEffectConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal item use effect config failed: %w", err)
	}

	cm.itemUseEffectConfigs = make(map[uint32]*ItemUseEffectConfig)
	for _, cfg := range configs {
		if cfg != nil {
			cm.itemUseEffectConfigs[cfg.ItemId] = cfg
		}
	}

	log.Infof("Loaded %d item use effect configs", len(cm.itemUseEffectConfigs))
	return nil
}

// loadItemRecycleConfigs 加载物品回收配置
func (cm *ConfigManager) loadItemRecycleConfigs() error {
	filePath := filepath.Join(cm.configPath, "item_recycle_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("item_recycle_config.json not found, using empty config")
			cm.itemRecycleConfigs = make(map[uint32]*ItemRecycleConfig)
			return nil
		}
		return fmt.Errorf("read item recycle config failed: %w", err)
	}

	var configs []*ItemRecycleConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal item recycle config failed: %w", err)
	}

	cm.itemRecycleConfigs = make(map[uint32]*ItemRecycleConfig)
	for _, cfg := range configs {
		if cfg != nil {
			cm.itemRecycleConfigs[cfg.ItemId] = cfg
		}
	}

	log.Infof("Loaded %d item recycle configs", len(cm.itemRecycleConfigs))
	return nil
}

// loadShopConfigs 加载商城配置
func (cm *ConfigManager) loadShopConfigs() error {
	filePath := filepath.Join(cm.configPath, "shop_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("shop_config.json not found, using empty config")
			cm.shopConfigs = make(map[uint32]*ShopConfig)
			return nil
		}
		return fmt.Errorf("read shop config failed: %w", err)
	}

	var configs []*ShopConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal shop config failed: %w", err)
	}

	cm.shopConfigs = make(map[uint32]*ShopConfig)
	for _, cfg := range configs {
		if cfg != nil {
			// 使用itemId作为key
			cm.shopConfigs[cfg.ItemId] = cfg
		}
	}

	log.Infof("Loaded %d shop configs", len(cm.shopConfigs))
	return nil
}

// GetItemUseEffectConfig 获取物品使用效果配置
func (cm *ConfigManager) GetItemUseEffectConfig(itemId uint32) (*ItemUseEffectConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.itemUseEffectConfigs[itemId]
	return cfg, ok
}

// GetItemRecycleConfig 获取物品回收配置
func (cm *ConfigManager) GetItemRecycleConfig(itemId uint32) (*ItemRecycleConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.itemRecycleConfigs[itemId]
	return cfg, ok
}

// GetShopConfig 获取商城配置（通过itemId）
func (cm *ConfigManager) GetShopConfig(itemId uint32) (*ShopConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.shopConfigs[itemId]
	return cfg, ok
}

// loadJobConfigs 加载职业配置
func (cm *ConfigManager) loadJobConfigs() error {
	filePath := filepath.Join(cm.configPath, "job_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("job_config.json not found, using empty config")
			cm.jobConfigs = make(map[uint32]*JobConfig)
			return nil
		}
		return fmt.Errorf("read job config failed: %w", err)
	}

	var configs []*JobConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal job config failed: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.jobConfigs = make(map[uint32]*JobConfig)
	for _, cfg := range configs {
		if cfg != nil {
			cm.jobConfigs[cfg.JobId] = cfg
		}
	}

	log.Infof("Loaded %d job configs", len(cm.jobConfigs))
	return nil
}

// loadSceneConfigs 加载场景配置
func (cm *ConfigManager) loadSceneConfigs() error {
	filePath := filepath.Join(cm.configPath, "scene_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("scene_config.json not found, using empty config")
			cm.sceneConfigs = make(map[uint32]*SceneConfig)
			return nil
		}
		return fmt.Errorf("read scene config failed: %w", err)
	}

	var configs []*SceneConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal scene config failed: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.sceneConfigs = make(map[uint32]*SceneConfig)
	for _, cfg := range configs {
		if cfg != nil {
			cm.sceneConfigs[cfg.SceneId] = cfg
		}
	}

	log.Infof("Loaded %d scene configs", len(cm.sceneConfigs))
	return nil
}

// loadMonsterSceneConfigs 加载怪物场景配置
func (cm *ConfigManager) loadMonsterSceneConfigs() error {
	filePath := filepath.Join(cm.configPath, "monster_scene_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("monster_scene_config.json not found, using empty config")
			cm.monsterSceneConfigs = make([]*MonsterSceneConfig, 0)
			return nil
		}
		return fmt.Errorf("read monster scene config failed: %w", err)
	}

	var configs []*MonsterSceneConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal monster scene config failed: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.monsterSceneConfigs = make([]*MonsterSceneConfig, 0, len(configs))
	for _, cfg := range configs {
		if cfg != nil {
			cm.monsterSceneConfigs = append(cm.monsterSceneConfigs, cfg)
		}
	}

	log.Infof("Loaded %d monster scene configs", len(cm.monsterSceneConfigs))
	return nil
}

// loadNPCSceneConfigs 加载NPC场景配置
func (cm *ConfigManager) loadNPCSceneConfigs() error {
	filePath := filepath.Join(cm.configPath, "npc_scene_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("npc_scene_config.json not found, using empty config")
			cm.npcSceneConfigs = make([]*NPCSceneConfig, 0)
			return nil
		}
		return fmt.Errorf("read npc scene config failed: %w", err)
	}

	var configs []*NPCSceneConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal npc scene config failed: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.npcSceneConfigs = make([]*NPCSceneConfig, 0, len(configs))
	for _, cfg := range configs {
		if cfg != nil {
			cm.npcSceneConfigs = append(cm.npcSceneConfigs, cfg)
		}
	}

	log.Infof("Loaded %d npc scene configs", len(cm.npcSceneConfigs))
	return nil
}

// loadTeleportConfigs 加载传送点配置
func (cm *ConfigManager) loadTeleportConfigs() error {
	filePath := filepath.Join(cm.configPath, "teleport_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("teleport_config.json not found, using empty config")
			cm.teleportConfigs = make(map[uint32]*TeleportConfig)
			return nil
		}
		return fmt.Errorf("read teleport config failed: %w", err)
	}

	var configs []*TeleportConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal teleport config failed: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.teleportConfigs = make(map[uint32]*TeleportConfig)
	for _, cfg := range configs {
		if cfg != nil {
			cm.teleportConfigs[cfg.TeleportId] = cfg
		}
	}

	log.Infof("Loaded %d teleport configs", len(cm.teleportConfigs))
	return nil
}

// GetJobConfig 获取职业配置
func (cm *ConfigManager) GetJobConfig(jobId uint32) (*JobConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.jobConfigs[jobId]
	return cfg, ok
}

// GetSceneConfig 获取场景配置
func (cm *ConfigManager) GetSceneConfig(sceneId uint32) (*SceneConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.sceneConfigs[sceneId]
	return cfg, ok
}

// GetMonsterSceneConfigs 获取指定场景的怪物配置列表
func (cm *ConfigManager) GetMonsterSceneConfigs(sceneId uint32) []*MonsterSceneConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]*MonsterSceneConfig, 0)
	for _, cfg := range cm.monsterSceneConfigs {
		if cfg != nil && cfg.SceneId == sceneId {
			result = append(result, cfg)
		}
	}
	return result
}

// GetNPCSceneConfig 根据NPC ID获取NPC配置
func (cm *ConfigManager) GetNPCSceneConfig(npcId uint32) *NPCSceneConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	for _, cfg := range cm.npcSceneConfigs {
		if cfg.NpcId == npcId {
			return cfg
		}
	}
	return nil
}

// GetNPCSceneConfigs 获取指定场景的NPC配置列表
func (cm *ConfigManager) GetNPCSceneConfigs(sceneId uint32) []*NPCSceneConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]*NPCSceneConfig, 0)
	for _, cfg := range cm.npcSceneConfigs {
		if cfg != nil && cfg.SceneId == sceneId {
			result = append(result, cfg)
		}
	}
	return result
}

// GetTeleportConfig 获取传送点配置
func (cm *ConfigManager) GetTeleportConfig(teleportId uint32) (*TeleportConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.teleportConfigs[teleportId]
	return cfg, ok
}

// GetTeleportConfigsByScene 获取指定场景的传送点配置列表
func (cm *ConfigManager) GetTeleportConfigsByScene(sceneId uint32) []*TeleportConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]*TeleportConfig, 0)
	for _, cfg := range cm.teleportConfigs {
		if cfg != nil && cfg.FromSceneId == sceneId {
			result = append(result, cfg)
		}
	}
	return result
}

// loadDailyActivityRewardConfigs 加载日常活跃奖励配置
func (cm *ConfigManager) loadDailyActivityRewardConfigs() error {
	filePath := filepath.Join(cm.configPath, "daily_activity_reward_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warnf("daily_activity_reward_config.json not found, using empty config")
			cm.dailyActivityRewardConfigs = make(map[uint32]*DailyActivityRewardConfig)
			return nil
		}
		return fmt.Errorf("read daily activity reward config failed: %w", err)
	}

	var configs []*DailyActivityRewardConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal daily activity reward config failed: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.dailyActivityRewardConfigs = make(map[uint32]*DailyActivityRewardConfig)
	for _, cfg := range configs {
		if cfg != nil {
			cm.dailyActivityRewardConfigs[cfg.RewardId] = cfg
		}
	}

	log.Infof("Loaded %d daily activity reward configs", len(cm.dailyActivityRewardConfigs))
	return nil
}

// GetDailyActivityRewardConfig 获取日常活跃奖励配置
func (cm *ConfigManager) GetDailyActivityRewardConfig(rewardId uint32) (*DailyActivityRewardConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.dailyActivityRewardConfigs[rewardId]
	return cfg, ok
}

// GetAllDailyActivityRewardConfigs 获取所有日常活跃奖励配置
func (cm *ConfigManager) GetAllDailyActivityRewardConfigs() []*DailyActivityRewardConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]*DailyActivityRewardConfig, 0, len(cm.dailyActivityRewardConfigs))
	for _, cfg := range cm.dailyActivityRewardConfigs {
		if cfg != nil {
			result = append(result, cfg)
		}
	}
	return result
}

// loadAchievementConfigs 加载成就配置
func (cm *ConfigManager) loadAchievementConfigs() error {
	filePath := filepath.Join(cm.configPath, "achievement_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("achievement_config.json not found, using empty config")
			cm.achievementConfigs = make(map[uint32]*AchievementConfig)
			return nil
		}
		return fmt.Errorf("read achievement config failed: %w", err)
	}

	var configs []*AchievementConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal achievement config failed: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.achievementConfigs = make(map[uint32]*AchievementConfig)
	for _, cfg := range configs {
		if cfg != nil {
			cm.achievementConfigs[cfg.AchievementId] = cfg
		}
	}

	log.Infof("Loaded %d achievement configs", len(cm.achievementConfigs))
	return nil
}

// GetAchievementConfig 获取成就配置
func (cm *ConfigManager) GetAchievementConfig(achievementId uint32) (*AchievementConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.achievementConfigs[achievementId]
	return cfg, ok
}

// GetAllAchievementConfigs 获取所有成就配置
func (cm *ConfigManager) GetAllAchievementConfigs() []*AchievementConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]*AchievementConfig, 0, len(cm.achievementConfigs))
	for _, cfg := range cm.achievementConfigs {
		if cfg != nil {
			result = append(result, cfg)
		}
	}
	return result
}

// loadEquipRefineConfigs 加载装备精炼配置
func (cm *ConfigManager) loadEquipRefineConfigs() error {
	filePath := filepath.Join(cm.configPath, "equip_refine_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("equip_refine_config.json not found, using empty config")
			cm.equipRefineConfigs = make(map[uint32]*EquipRefineConfig)
			return nil
		}
		return fmt.Errorf("read equip refine config failed: %w", err)
	}

	var configs []*EquipRefineConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal equip refine config failed: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.equipRefineConfigs = make(map[uint32]*EquipRefineConfig)
	for _, cfg := range configs {
		if cfg != nil {
			cm.equipRefineConfigs[cfg.ItemId] = cfg
		}
	}

	log.Infof("Loaded %d equip refine configs", len(cm.equipRefineConfigs))
	return nil
}

// GetEquipRefineConfig 获取装备精炼配置
func (cm *ConfigManager) GetEquipRefineConfig(itemId uint32) (*EquipRefineConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.equipRefineConfigs[itemId]
	return cfg, ok
}

// loadEquipEnchantConfigs 加载装备附魔配置
func (cm *ConfigManager) loadEquipEnchantConfigs() error {
	filePath := filepath.Join(cm.configPath, "equip_enchant_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("equip_enchant_config.json not found, using empty config")
			cm.equipEnchantConfigs = make(map[uint32]*EquipEnchantConfig)
			return nil
		}
		return fmt.Errorf("read equip enchant config failed: %w", err)
	}

	var configs []*EquipEnchantConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal equip enchant config failed: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.equipEnchantConfigs = make(map[uint32]*EquipEnchantConfig)
	for _, cfg := range configs {
		if cfg != nil {
			cm.equipEnchantConfigs[cfg.ItemId] = cfg
		}
	}

	log.Infof("Loaded %d equip enchant configs", len(cm.equipEnchantConfigs))
	return nil
}

// GetEquipEnchantConfig 获取装备附魔配置
func (cm *ConfigManager) GetEquipEnchantConfig(itemId uint32) (*EquipEnchantConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.equipEnchantConfigs[itemId]
	return cfg, ok
}

// loadEquipSetConfigs 加载装备套装配置
func (cm *ConfigManager) loadEquipSetConfigs() error {
	filePath := filepath.Join(cm.configPath, "equip_set_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("equip_set_config.json not found, using empty config")
			cm.equipSetConfigs = make(map[uint32]*EquipSetConfig)
			return nil
		}
		return fmt.Errorf("read equip set config failed: %w", err)
	}

	var configs []*EquipSetConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal equip set config failed: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.equipSetConfigs = make(map[uint32]*EquipSetConfig)
	for _, cfg := range configs {
		if cfg != nil {
			cm.equipSetConfigs[cfg.SetId] = cfg
		}
	}

	log.Infof("Loaded %d equip set configs", len(cm.equipSetConfigs))
	return nil
}

// GetEquipSetConfig 获取装备套装配置
func (cm *ConfigManager) GetEquipSetConfig(setId uint32) (*EquipSetConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.equipSetConfigs[setId]
	return cfg, ok
}

// loadBagConfigs 加载背包配置
func (cm *ConfigManager) loadBagConfigs() error {
	filePath := filepath.Join(cm.configPath, "bag_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("bag_config.json not found, using empty config")
			cm.mu.Lock()
			cm.bagConfigs = make(map[uint32]*BagConfig)
			cm.mu.Unlock()
			return nil
		}
		return fmt.Errorf("read bag config failed: %w", err)
	}

	var configs []*BagConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal bag config failed: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.bagConfigs = make(map[uint32]*BagConfig)
	for _, cfg := range configs {
		if cfg != nil {
			cm.bagConfigs[cfg.BagType] = cfg
		}
	}

	log.Infof("Loaded %d bag configs", len(cm.bagConfigs))
	return nil
}

// GetBagConfig 获取背包配置（通过bagType）
func (cm *ConfigManager) GetBagConfig(bagType uint32) (*BagConfig, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cfg, ok := cm.bagConfigs[bagType]
	return cfg, ok
}
