package jsonconf

import (
	"os"
	"path/filepath"
	"sync"

	"postapocgame/server/internal"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/pkg/log"
)

type AttrPushConfig struct {
	PushAll            bool     `json:"push_all"`
	IncludeSysPowerMap bool     `json:"include_sys_power_map"`
	Systems            []uint32 `json:"systems"`
}

type AttrPowerWeight struct {
	AttrWeights map[uint32]float64 `json:"attr_weights"`
}

type AttrPowerConfig struct {
	Default AttrPowerWeight            `json:"default"`
	Jobs    map[uint32]AttrPowerWeight `json:"jobs"`
}

type AttrAddRateConfig struct {
	Level struct {
		HPRegenPerLevel int64 `json:"hp_regen_per_level"`
		MPRegenPerLevel int64 `json:"mp_regen_per_level"`
	} `json:"level"`
	Equip struct {
		DamageAddPerRefine int64 `json:"damage_add_per_refine"`
	} `json:"equip"`
}

type AttrFormulaConversion struct {
	From attrdef.AttrType   `json:"from"`
	To   []attrdef.AttrType `json:"to"`
}

type AttrFormulaPercentage struct {
	RateAttr attrdef.AttrType   `json:"rate"`
	Targets  []attrdef.AttrType `json:"targets"`
	Mode     string             `json:"mode,omitempty"`
}

type AttrFormulaConfig struct {
	Conversions []AttrFormulaConversion `json:"conversions"`
	Percentages []AttrFormulaPercentage `json:"percentages"`
}

type AttrConfig struct {
	Push    AttrPushConfig    `json:"push"`
	Power   AttrPowerConfig   `json:"power"`
	AddRate AttrAddRateConfig `json:"add_rate"`
	Formula AttrFormulaConfig `json:"formula"`
}

var (
	attrCfgOnce sync.Once
)

func defaultAttrPushConfig() AttrPushConfig {
	return AttrPushConfig{
		PushAll:            true,
		IncludeSysPowerMap: true,
		Systems:            []uint32{},
	}
}

func defaultAttrPowerConfig() AttrPowerConfig {
	return AttrPowerConfig{
		Default: AttrPowerWeight{
			AttrWeights: map[uint32]float64{
				uint32(attrdef.AttrAttack):    2.0,
				uint32(attrdef.AttrDefense):   1.5,
				uint32(attrdef.AttrMaxHP):     0.5,
				uint32(attrdef.AttrCritRate):  0.8,
				uint32(attrdef.AttrDamageAdd): 1.0,
			},
		},
		Jobs: make(map[uint32]AttrPowerWeight),
	}
}

func defaultAttrAddRateConfig() AttrAddRateConfig {
	cfg := AttrAddRateConfig{}
	cfg.Level.HPRegenPerLevel = 50
	cfg.Level.MPRegenPerLevel = 50
	cfg.Equip.DamageAddPerRefine = 25
	return cfg
}

func defaultAttrFormulaConfig() AttrFormulaConfig {
	return AttrFormulaConfig{
		Conversions: []AttrFormulaConversion{
			{From: attrdef.AttrAttack, To: []attrdef.AttrType{attrdef.AttrAttackPhysical, attrdef.AttrAttackMagic}},
			{From: attrdef.AttrDefense, To: []attrdef.AttrType{attrdef.AttrDefensePhysical, attrdef.AttrDefenseMagic}},
		},
		Percentages: []AttrFormulaPercentage{
			{RateAttr: attrdef.AttrDamageAdd, Targets: []attrdef.AttrType{attrdef.AttrAttack, attrdef.AttrAttackPhysical, attrdef.AttrAttackMagic}},
			{RateAttr: attrdef.AttrDamageReduce, Targets: []attrdef.AttrType{attrdef.AttrDefense, attrdef.AttrDefensePhysical, attrdef.AttrDefenseMagic}},
			{RateAttr: attrdef.AttrMaxHPAddRate, Targets: []attrdef.AttrType{attrdef.AttrMaxHP}},
			{RateAttr: attrdef.AttrAttackAddRate, Targets: []attrdef.AttrType{attrdef.AttrAttack, attrdef.AttrAttackPhysical, attrdef.AttrAttackMagic}},
			{RateAttr: attrdef.AttrDefenseAddRate, Targets: []attrdef.AttrType{attrdef.AttrDefense, attrdef.AttrDefensePhysical, attrdef.AttrDefenseMagic}},
			{RateAttr: attrdef.AttrSpeedAddRate, Targets: []attrdef.AttrType{attrdef.AttrSpeed}},
		},
	}
}

func defaultAttrConfig() *AttrConfig {
	return &AttrConfig{
		Push:    defaultAttrPushConfig(),
		Power:   defaultAttrPowerConfig(),
		AddRate: defaultAttrAddRateConfig(),
		Formula: defaultAttrFormulaConfig(),
	}
}

func (cm *ConfigManager) loadAttrConfigs() {
	attrCfgOnce.Do(func() {
		path := filepath.Join(cm.configPath, "attr_config.json")
		data, err := os.ReadFile(path)
		if err != nil {
			log.Warnf("attr_config: load failed (%v), using defaults", err)
			cm.attrConfig = defaultAttrConfig()
			return
		}
		cfg := &AttrConfig{}
		if err := internal.Unmarshal(data, cfg); err != nil {
			log.Warnf("attr_config: parse failed (%v), using defaults", err)
			cm.attrConfig = defaultAttrConfig()
			return
		}
		cm.attrConfig = cfg
	})
}

func (cm *ConfigManager) ensureAttrConfig() {
	if cm == nil {
		return
	}
	if cm.attrConfig == nil {
		cm.attrConfig = &AttrConfig{}
		cm.loadAttrConfigs()
	}
}

func (cm *ConfigManager) GetAttrPushConfig() *AttrPushConfig {
	cm.ensureAttrConfig()
	if cm == nil || cm.attrConfig == nil {
		return nil
	}
	return &cm.attrConfig.Push
}

func (cm *ConfigManager) GetAttrPowerConfig() *AttrPowerConfig {
	cm.ensureAttrConfig()
	if cm == nil || cm.attrConfig == nil {
		return nil
	}
	return &cm.attrConfig.Power
}

func (cm *ConfigManager) GetAttrAddRateConfig() *AttrAddRateConfig {
	cm.ensureAttrConfig()
	if cm == nil || cm.attrConfig == nil {
		return nil
	}
	return &cm.attrConfig.AddRate
}

func (cm *ConfigManager) GetAttrFormulaConfig() *AttrFormulaConfig {
	cm.ensureAttrConfig()
	if cm == nil || cm.attrConfig == nil {
		return nil
	}
	return &cm.attrConfig.Formula
}
