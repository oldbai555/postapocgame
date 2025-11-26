package attrpower

import (
	"math"

	icalc "postapocgame/server/internal/attrcalc"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/jsonconf"
)

// CalcPower 计算战力
func CalcPower(calc *icalc.FightAttrCalc, job uint32) int64 {
	if calc == nil {
		return 0
	}
	weights := getWeights(job)
	if len(weights.AttrWeights) == 0 {
		return 0
	}
	var value float64
	calc.DoRange(func(attrType attrdef.AttrType, attrValue attrdef.AttrValue) {
		if w, ok := weights.AttrWeights[uint32(attrType)]; ok && w > 0 {
			value += float64(attrValue) * w
		}
	})
	return int64(math.Round(value))
}

func getWeights(job uint32) jsonconf.AttrPowerWeight {
	cfg := load()
	if cfg == nil {
		return jsonconf.AttrPowerWeight{}
	}
	if w, ok := cfg.Jobs[job]; ok {
		return mergeWeights(cfg.Default, w)
	}
	return cfg.Default
}

func load() *jsonconf.AttrPowerConfig {
	return jsonconf.GetConfigManager().GetAttrPowerConfig()
}

func mergeWeights(base jsonconf.AttrPowerWeight, override jsonconf.AttrPowerWeight) jsonconf.AttrPowerWeight {
	result := jsonconf.AttrPowerWeight{
		AttrWeights: make(map[uint32]float64, len(base.AttrWeights)+len(override.AttrWeights)),
	}
	for k, v := range base.AttrWeights {
		result.AttrWeights[k] = v
	}
	for k, v := range override.AttrWeights {
		result.AttrWeights[k] = v
	}
	return result
}
