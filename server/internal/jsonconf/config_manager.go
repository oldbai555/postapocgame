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
	skillConfigs map[uint32]*SkillConfig
	jobConfigs   map[uint32]*JobConfig
	sceneConfigs map[uint32]*SceneConfig
	mapConfigs   map[uint32]*MapConfig
}

var (
	globalConfigManager = newConfigManager()
	configOnce          sync.Once
)

func newConfigManager() *ConfigManager {
	return &ConfigManager{
		skillConfigs: make(map[uint32]*SkillConfig),
		jobConfigs:   make(map[uint32]*JobConfig),
		sceneConfigs: make(map[uint32]*SceneConfig),
		mapConfigs:   make(map[uint32]*MapConfig),
	}
}

// GetConfigManager 获取全局配置管理器
func GetConfigManager() *ConfigManager {
	configOnce.Do(func() {
		globalConfigManager = newConfigManager()
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

	// 加载技能配置
	if err := cm.loadSkillConfigs(); err != nil {
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

	// 加载地图配置
	if err := cm.loadMapConfigs(); err != nil {
		return customerr.Wrap(err)
	}

	log.Infof("All configs loaded successfully")
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

// Reload 热加载配置
func (cm *ConfigManager) Reload() error {
	log.Infof("Reloading all configs...")
	return cm.LoadAllConfigs()
}

// GetSkillConfig 获取技能配置，未找到返回 nil
func (cm *ConfigManager) GetSkillConfig(skillId uint32) *SkillConfig {
	if cm == nil || cm.skillConfigs == nil {
		return nil
	}
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.skillConfigs[skillId]
}

// loadJobConfigs 加载职业配置
func (cm *ConfigManager) loadJobConfigs() error {
	filePath := filepath.Join(cm.configPath, "jobconfig.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read job config failed: %w", err)
	}

	var configs []*JobConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal job config failed: %w", err)
	}

	// 注意：LoadAllConfigs 已经持有锁，这里不需要再次获取锁
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
	filePath := filepath.Join(cm.configPath, "sceneconfig.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 如果文件不存在，使用空配置（允许可选）
		if os.IsNotExist(err) {
			log.Warnf("sceneconfig.json not found, using empty config")
			cm.sceneConfigs = make(map[uint32]*SceneConfig)
			return nil
		}
		return fmt.Errorf("read scene config failed: %w", err)
	}

	var configs []*SceneConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal scene config failed: %w", err)
	}

	// 注意：LoadAllConfigs 已经持有锁，这里不需要再次获取锁
	cm.sceneConfigs = make(map[uint32]*SceneConfig)
	for _, cfg := range configs {
		if cfg != nil {
			cm.sceneConfigs[cfg.SceneId] = cfg
		}
	}

	log.Infof("Loaded %d scene configs", len(cm.sceneConfigs))
	return nil
}

// loadMapConfigs 加载地图配置
func (cm *ConfigManager) loadMapConfigs() error {
	filePath := filepath.Join(cm.configPath, "mapconfig.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warnf("mapconfig.json not found, skip map binding")
			// 注意：LoadAllConfigs 已经持有锁，这里不需要再次获取锁
			cm.mapConfigs = make(map[uint32]*MapConfig)
			cm.bindMapsToScenesLocked()
			return nil
		}
		return fmt.Errorf("read map config failed: %w", err)
	}

	var configs []*MapConfig
	if err := internal.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("unmarshal map config failed: %w", err)
	}

	prepared := make(map[uint32]*MapConfig)
	for _, cfg := range configs {
		if cfg == nil {
			continue
		}
		gameMap, err := newGameMapFromTileData(cfg.TileData)
		if err != nil {
			return fmt.Errorf("mapId=%d invalid: %w", cfg.MapId, err)
		}
		cfg.gameMap = gameMap
		prepared[cfg.MapId] = cfg
	}

	// 注意：LoadAllConfigs 已经持有锁，这里不需要再次获取锁
	cm.mapConfigs = prepared
	bound := cm.bindMapsToScenesLocked()

	log.Infof("Loaded %d map configs, bound to %d scenes", len(cm.mapConfigs), bound)
	return nil
}

func (cm *ConfigManager) bindMapsToScenesLocked() int {
	bound := 0
	for _, sceneCfg := range cm.sceneConfigs {
		if sceneCfg == nil {
			continue
		}
		sceneCfg.GameMap = nil
		if sceneCfg.MapId == 0 {
			continue
		}
		mapCfg, ok := cm.mapConfigs[sceneCfg.MapId]
		if !ok || mapCfg == nil || mapCfg.gameMap == nil {
			log.Warnf("scene %d references missing mapId=%d", sceneCfg.SceneId, sceneCfg.MapId)
			continue
		}
		sceneCfg.GameMap = mapCfg.gameMap
		sceneCfg.Width = int(sceneCfg.GameMap.Width())
		sceneCfg.Height = int(sceneCfg.GameMap.Height())
		bound++
	}
	return bound
}

// GetJobConfig 获取职业配置，未找到返回 nil
func (cm *ConfigManager) GetJobConfig(jobId uint32) *JobConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.jobConfigs[jobId]
}

// GetSceneConfig 获取场景配置
func (cm *ConfigManager) GetSceneConfig(sceneId uint32) *SceneConfig {
	if cm == nil || cm.sceneConfigs == nil {
		return nil
	}
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.sceneConfigs[sceneId]
}

// GetMapConfig 获取地图配置
func (cm *ConfigManager) GetMapConfig(mapId uint32) *MapConfig {
	if cm == nil || cm.mapConfigs == nil {
		return nil
	}
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.mapConfigs[mapId]
}
