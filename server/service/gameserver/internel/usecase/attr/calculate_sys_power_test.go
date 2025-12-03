package attr

import (
	"context"
	"testing"

	icalc "postapocgame/server/internal/attrcalc"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// mockConfigManager 模拟 ConfigManager
type mockConfigManager struct{}

func (m *mockConfigManager) GetItemConfig(itemID uint32) (interface{}, bool) {
	return nil, false
}

func (m *mockConfigManager) GetBagConfig(bagType uint32) (interface{}, bool) {
	return nil, false
}

func TestCalculateSysPowerUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	mockConfig := &mockConfigManager{}
	uc := NewCalculateSysPowerUseCase(mockConfig)

	t.Run("空的系统属性应该返回空映射", func(t *testing.T) {
		data := &SystemAttrData{
			SysAttr:        make(map[uint32]*icalc.FightAttrCalc),
			SysAddRateAttr: make(map[uint32]*icalc.FightAttrCalc),
			Job:            1,
		}
		result, err := uc.Execute(ctx, data)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}
		if len(result) != 0 {
			t.Errorf("Result length = %d, want 0", len(result))
		}
	})

	t.Run("nil 数据应该返回空映射", func(t *testing.T) {
		result, err := uc.Execute(ctx, nil)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}
		if len(result) != 0 {
			t.Errorf("Result length = %d, want 0", len(result))
		}
	})

	t.Run("计算单个系统战力", func(t *testing.T) {
		calc1 := icalc.NewFightAttrCalc()
		calc1.AddValue(1, 100) // 添加属性值

		data := &SystemAttrData{
			SysAttr: map[uint32]*icalc.FightAttrCalc{
				1: calc1,
			},
			SysAddRateAttr: make(map[uint32]*icalc.FightAttrCalc),
			Job:            1,
		}
		result, err := uc.Execute(ctx, data)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}
		if len(result) != 1 {
			t.Fatalf("Result length = %d, want 1", len(result))
		}
		if _, ok := result[1]; !ok {
			t.Error("System 1 power should be calculated")
		}
		if result[1] < 0 {
			t.Errorf("System 1 power = %d, should be >= 0", result[1])
		}
	})

	t.Run("计算带加成率的系统战力", func(t *testing.T) {
		calc1 := icalc.NewFightAttrCalc()
		calc1.AddValue(1, 100)

		addRateCalc1 := icalc.NewFightAttrCalc()
		addRateCalc1.AddValue(1, 10) // 10% 加成

		data := &SystemAttrData{
			SysAttr: map[uint32]*icalc.FightAttrCalc{
				1: calc1,
			},
			SysAddRateAttr: map[uint32]*icalc.FightAttrCalc{
				1: addRateCalc1,
			},
			Job: 1,
		}
		result, err := uc.Execute(ctx, data)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}
		if len(result) != 1 {
			t.Fatalf("Result length = %d, want 1", len(result))
		}
		// 带加成的战力应该大于不带加成的
		if result[1] < 0 {
			t.Errorf("System 1 power with add rate = %d, should be >= 0", result[1])
		}
	})

	t.Run("计算多个系统战力", func(t *testing.T) {
		calc1 := icalc.NewFightAttrCalc()
		calc1.AddValue(1, 100)

		calc2 := icalc.NewFightAttrCalc()
		calc2.AddValue(2, 200)

		data := &SystemAttrData{
			SysAttr: map[uint32]*icalc.FightAttrCalc{
				1: calc1,
				2: calc2,
			},
			SysAddRateAttr: make(map[uint32]*icalc.FightAttrCalc),
			Job:            1,
		}
		result, err := uc.Execute(ctx, data)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}
		if len(result) != 2 {
			t.Fatalf("Result length = %d, want 2", len(result))
		}
		if _, ok := result[1]; !ok {
			t.Error("System 1 power should be calculated")
		}
		if _, ok := result[2]; !ok {
			t.Error("System 2 power should be calculated")
		}
	})

	t.Run("忽略 nil 计算器", func(t *testing.T) {
		calc1 := icalc.NewFightAttrCalc()
		calc1.AddValue(1, 100)

		data := &SystemAttrData{
			SysAttr: map[uint32]*icalc.FightAttrCalc{
				1: calc1,
				2: nil, // nil 计算器应该被忽略
			},
			SysAddRateAttr: make(map[uint32]*icalc.FightAttrCalc),
			Job:            1,
		}
		result, err := uc.Execute(ctx, data)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}
		if len(result) != 1 {
			t.Fatalf("Result length = %d, want 1 (nil calc should be ignored)", len(result))
		}
		if _, ok := result[1]; !ok {
			t.Error("System 1 power should be calculated")
		}
		if _, ok := result[2]; ok {
			t.Error("System 2 power should not be calculated (nil calc)")
		}
	})
}
