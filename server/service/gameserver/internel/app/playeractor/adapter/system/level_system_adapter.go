package system

import (
	"context"
	icalc "postapocgame/server/internal/attrcalc"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/system/attrcalc"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	level2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/level"
	money2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/money"
	reward2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/reward"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

// LevelSystemAdapter 等级系统适配器
//
// 生命周期职责：
// - OnInit: 调用 InitLevelDataUseCase 初始化等级数据（默认值修正、经验同步）
// - 其他生命周期: 暂未使用
//
// 业务逻辑：所有业务逻辑（加经验、升级、属性加成）均在 UseCase 层实现
//
// ⚠️ 防退化机制：禁止在 SystemAdapter 中编写业务规则逻辑，只允许调用 UseCase 与管理生命周期

// EnsureISystem 确保 LevelSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*LevelSystemAdapter)(nil)

type LevelSystemAdapter struct {
	*BaseSystemAdapter
	addExpUseCase        *level2.AddExpUseCase
	levelUpUseCase       *level2.LevelUpUseCase
	initLevelDataUseCase *level2.InitLevelDataUseCase
}

// NewLevelSystemAdapter 创建等级系统适配器
func NewLevelSystemAdapter() *LevelSystemAdapter {
	rewardUseCase := reward2.NewRewardUseCase(deps.PlayerGateway(), deps.EventPublisher(), deps.ConfigGateway())
	moneyUseCase := money2.NewMoneyUseCaseImpl(deps.PlayerGateway())
	levelUpUC := level2.NewLevelUpUseCase(
		deps.PlayerGateway(),
		deps.EventPublisher(),
		deps.ConfigGateway(),
	)

	// 注入 AttrUseCase 依赖（通过适配器）
	attrUseCase := NewAttrUseCaseAdapter()
	levelUpUC.SetDependencies(moneyUseCase, attrUseCase, rewardUseCase)

	addExpUC := level2.NewAddExpUseCase(
		deps.PlayerGateway(),
		deps.EventPublisher(),
		deps.ConfigGateway(),
	)
	addExpUC.SetDependencies(moneyUseCase, attrUseCase, rewardUseCase)

	return &LevelSystemAdapter{
		BaseSystemAdapter:    NewBaseSystemAdapter(uint32(protocol.SystemId_SysLevel)),
		addExpUseCase:        addExpUC,
		levelUpUseCase:       levelUpUC,
		initLevelDataUseCase: level2.NewInitLevelDataUseCase(deps.PlayerGateway()),
	}
}

// OnInit 系统初始化
func (a *LevelSystemAdapter) OnInit(ctx context.Context) {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("level sys OnInit get role err:%v", err)
		return
	}
	// 初始化等级数据（包括默认值修正、经验同步等业务逻辑）
	if err := a.initLevelDataUseCase.Execute(ctx, roleID); err != nil {
		log.Errorf("level sys OnInit init level data err:%v", err)
		return
	}
}

// AddExp 添加经验值（对外接口，供其他系统调用）
func (a *LevelSystemAdapter) AddExp(ctx context.Context, exp uint64) error {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	return a.addExpUseCase.Execute(ctx, roleID, exp)
}

// GetLevel 获取当前等级
func (a *LevelSystemAdapter) GetLevel(ctx context.Context) (uint32, error) {
	levelData, err := deps.PlayerGateway().GetLevelData(ctx)
	if err != nil {
		return 0, err
	}
	return levelData.Level, nil
}

// GetExp 获取当前经验
func (a *LevelSystemAdapter) GetExp(ctx context.Context) (int64, error) {
	levelData, err := deps.PlayerGateway().GetLevelData(ctx)
	if err != nil {
		return 0, err
	}
	return levelData.Exp, nil
}

// GetLevelData 获取等级数据
func (a *LevelSystemAdapter) GetLevelData(ctx context.Context) (*protocol.SiLevelData, error) {
	levelData, err := deps.PlayerGateway().GetLevelData(ctx)
	if err != nil {
		return nil, err
	}
	return levelData, nil
}

// CalculateAttrs 计算等级系统的属性（实现属性计算器接口）
func (a *LevelSystemAdapter) CalculateAttrs(ctx context.Context) []*protocol.AttrSt {
	levelData, err := deps.PlayerGateway().GetLevelData(ctx)
	if err != nil {
		return nil
	}
	// 从配置表获取等级属性
	levelAttrs := jsonconf.GetConfigManager().GetLevelAttrs(levelData.Level)
	if len(levelAttrs) == 0 {
		return nil
	}

	// 转换为protocol.AttrSt格式
	result := make([]*protocol.AttrSt, 0, len(levelAttrs))
	for attrType, attrValue := range levelAttrs {
		result = append(result, &protocol.AttrSt{
			Type:  attrType,
			Value: int64(attrValue),
		})
	}

	return result
}

// CalculateAddRate 计算等级加成属性
func (a *LevelSystemAdapter) CalculateAddRate(ctx context.Context, _ *icalc.FightAttrCalc) []*protocol.AttrSt {
	levelSys := GetLevelSys(ctx)
	if levelSys == nil {
		return nil
	}
	level, err := levelSys.GetLevel(ctx)
	if err != nil || level == 0 {
		return nil
	}
	cfg := jsonconf.GetConfigManager().GetAttrAddRateConfig()
	if cfg == nil {
		return nil
	}
	results := make([]*protocol.AttrSt, 0, 2)
	if cfg.Level.HPRegenPerLevel > 0 {
		regenHP := int64(level) * cfg.Level.HPRegenPerLevel
		if regenHP > 0 {
			results = append(results, &protocol.AttrSt{
				Type:  uint32(attrdef.AttrHPRegen),
				Value: regenHP,
			})
		}
	}
	if cfg.Level.MPRegenPerLevel > 0 {
		regenMP := int64(level) * cfg.Level.MPRegenPerLevel
		if regenMP > 0 {
			results = append(results, &protocol.AttrSt{
				Type:  uint32(attrdef.AttrMPRegen),
				Value: regenMP,
			})
		}
	}
	if len(results) == 0 {
		return nil
	}
	return results
}

// GetLevelSys 获取等级系统
func GetLevelSys(ctx context.Context) *LevelSystemAdapter {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysLevel))
	if system == nil {
		log.Errorf("not found system [%v]", protocol.SystemId_SysLevel)
		return nil
	}
	sys, ok := system.(*LevelSystemAdapter)
	if !ok {
		log.Errorf("invalid system type for [%v]", protocol.SystemId_SysLevel)
		return nil
	}
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error", protocol.SystemId_SysLevel)
		return nil
	}
	return sys
}

// 注册系统工厂和属性计算器
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysLevel), func() iface.ISystem {
		return NewLevelSystemAdapter()
	})

	// 注册属性计算器
	attrcalc.Register(uint32(protocol.SaAttrSys_SaLevel), func(ctx context.Context) attrcalc.Calculator {
		return GetLevelSys(ctx)
	})

	// 注册属性加成计算器
	attrcalc.RegisterAddRate(uint32(protocol.SaAttrSys_SaLevel), func(ctx context.Context) attrcalc.AddRateCalculator {
		return GetLevelSys(ctx)
	})
}
