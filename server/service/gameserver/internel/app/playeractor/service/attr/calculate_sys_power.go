package attr

import (
	"context"
	icalc "postapocgame/server/internal/attrcalc"
	"postapocgame/server/internal/attrpower"
	"postapocgame/server/service/gameserver/internel/iface"
)

// CalculateSysPowerUseCase 计算系统战力用例
// 负责根据系统属性计算各系统的战力值
type CalculateSysPowerUseCase struct {
	configManager iface.ConfigManager
}

// NewCalculateSysPowerUseCase 创建计算系统战力用例
func NewCalculateSysPowerUseCase(configManager iface.ConfigManager) *CalculateSysPowerUseCase {
	return &CalculateSysPowerUseCase{
		configManager: configManager,
	}
}

// SystemAttrData 系统属性数据（用于计算战力）
type SystemAttrData struct {
	SysAttr        map[uint32]*icalc.FightAttrCalc
	SysAddRateAttr map[uint32]*icalc.FightAttrCalc
	Job            uint32
}

// Execute 执行计算系统战力用例
// 返回各系统的战力映射
func (uc *CalculateSysPowerUseCase) Execute(ctx context.Context, data *SystemAttrData) (map[uint32]int64, error) {
	if data == nil || len(data.SysAttr) == 0 {
		return make(map[uint32]int64), nil
	}

	sysPower := make(map[uint32]int64, len(data.SysAttr))
	for sysID, calc := range data.SysAttr {
		if calc == nil {
			continue
		}
		// 创建临时计算器，合并基础属性和加成率
		temp := icalc.NewFightAttrCalc()
		temp.AddCalc(calc)
		if addRateCalc := data.SysAddRateAttr[sysID]; addRateCalc != nil {
			temp.AddCalc(addRateCalc)
		}
		// 应用转换和百分比计算
		icalc.ApplyConversions(temp)
		icalc.ApplyPercentages(temp)
		// 计算战力
		sysPower[sysID] = attrpower.CalcPower(temp, data.Job)
	}

	return sysPower, nil
}
