package attrcalc

import "postapocgame/server/internal/jsonconf"

func loadFormulaConfig() *jsonconf.AttrFormulaConfig {
	return jsonconf.GetConfigManager().GetAttrFormulaConfig()
}

// ApplyConversions 根据配置复制属性
func ApplyConversions(calc *FightAttrCalc) {
	if calc == nil {
		return
	}
	cfg := loadFormulaConfig()
	if cfg == nil {
		return
	}
	for _, rule := range cfg.Conversions {
		if len(rule.To) == 0 {
			continue
		}
		value := calc.GetValue(rule.From)
		if value == 0 {
			continue
		}
		for _, target := range rule.To {
			if target == 0 {
				continue
			}
			calc.AddValue(target, value)
		}
	}
}

// ApplyPercentages 按配置对属性应用百分比加成（以万分比为基数）
func ApplyPercentages(calc *FightAttrCalc) {
	if calc == nil {
		return
	}
	cfg := loadFormulaConfig()
	if cfg == nil {
		return
	}
	for _, rule := range cfg.Percentages {
		if rule.RateAttr == 0 || len(rule.Targets) == 0 {
			continue
		}
		rate := calc.GetValue(rule.RateAttr)
		if rate == 0 {
			continue
		}
		for _, target := range rule.Targets {
			if target == 0 {
				continue
			}
			base := calc.GetValue(target)
			if base == 0 {
				continue
			}
			delta := base * rate / 10000
			if delta == 0 {
				continue
			}
			switch rule.Mode {
			case "override":
				calc.SetValue(target, delta)
			default:
				calc.AddValue(target, delta)
			}
		}
	}
}
