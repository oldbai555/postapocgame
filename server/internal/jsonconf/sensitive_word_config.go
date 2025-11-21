package jsonconf

// SensitiveWordConfig 敏感词配置
type SensitiveWordConfig struct {
	Words []string `json:"words"` // 敏感词列表
}

// GetSensitiveWordConfig 获取敏感词配置
func (cm *ConfigManager) GetSensitiveWordConfig() *SensitiveWordConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.sensitiveWordConfig
}

// loadSensitiveWordConfig 加载敏感词配置
func (cm *ConfigManager) loadSensitiveWordConfig() error {
	// 如果文件不存在，使用默认敏感词列表
	config := &SensitiveWordConfig{
		Words: []string{
			"测试敏感词1",
			"测试敏感词2",
		},
	}
	cm.sensitiveWordConfig = config
	return nil
}
