package level

import (
	"context"
	icalc "postapocgame/server/internal/attrcalc"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/attrcalc"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/app/playeractor/money"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	level2 "postapocgame/server/service/gameserver/internel/app/playeractor/service/level"
	reward2 "postapocgame/server/service/gameserver/internel/app/playeractor/service/reward"
	"postapocgame/server/service/gameserver/internel/app/playeractor/sysbase"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

// Deps 聚合 Level 系统所需依赖，便于后续从 Runtime 或其他入口统一装配
type Deps struct {
	PlayerRepo     iface.PlayerRepository
	EventPublisher iface.EventPublisher
	ConfigManager  iface.ConfigManager
	NetworkGateway iface.NetworkGateway
}

// depsFromRuntime 从 Runtime 组装 Level 所需依赖
func depsFromRuntime(rt *runtime.Runtime) Deps {
	return Deps{
		PlayerRepo:     rt.PlayerRepo(),
		EventPublisher: rt.EventPublisher(),
		ConfigManager:  rt.ConfigManager(),
		NetworkGateway: rt.NetworkGateway(),
	}
}

var _ iface.ISystem = (*LevelSystemAdapter)(nil)

// LevelSystemAdapter 等级系统适配器
type LevelSystemAdapter struct {
	*sysbase.BaseSystem
	deps                 Deps
	addExpUseCase        *level2.AddExpUseCase
	levelUpUseCase       *level2.LevelUpUseCase
	initLevelDataUseCase *level2.InitLevelDataUseCase
}

// NewLevelSystemAdapter 创建等级系统适配器
// 如果 rt 为 nil，则使用 deps 工厂函数作为回退（兼容旧代码）
func NewLevelSystemAdapter(rt *runtime.Runtime) *LevelSystemAdapter {
	var levelDeps Deps
	if rt != nil {
		levelDeps = depsFromRuntime(rt)
	} else {
		// 回退到使用 deps 工厂函数（兼容旧代码）
		levelDeps = Deps{
			PlayerRepo:     deps.NewPlayerGateway(),
			EventPublisher: deps.NewEventPublisher(),
			ConfigManager:  deps.NewConfigManager(),
			NetworkGateway: deps.NewNetworkGateway(),
		}
	}

	rewardUseCase := reward2.NewRewardUseCase(levelDeps.PlayerRepo, levelDeps.EventPublisher, levelDeps.ConfigManager)
	moneyDeps := money.Deps{
		PlayerRepo:     levelDeps.PlayerRepo,
		EventPublisher: levelDeps.EventPublisher,
		NetworkGateway: levelDeps.NetworkGateway,
	}
	moneyUseCase := money.NewMoneyUseCaseImpl(moneyDeps)
	levelUpUC := level2.NewLevelUpUseCase(
		levelDeps.PlayerRepo,
		levelDeps.EventPublisher,
		levelDeps.ConfigManager,
	)

	// 注入依赖（属性系统改为工具类，这里仅保留货币与奖励用例）
	levelUpUC.SetDependencies(moneyUseCase, rewardUseCase)

	addExpUC := level2.NewAddExpUseCase(
		levelDeps.PlayerRepo,
		levelDeps.EventPublisher,
		levelDeps.ConfigManager,
	)
	addExpUC.SetDependencies(moneyUseCase, rewardUseCase)

	return &LevelSystemAdapter{
		BaseSystem:           sysbase.NewBaseSystem(uint32(protocol.SystemId_SysLevel)),
		deps:                 levelDeps,
		addExpUseCase:        addExpUC,
		levelUpUseCase:       levelUpUC,
		initLevelDataUseCase: level2.NewInitLevelDataUseCase(levelDeps.PlayerRepo),
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
	levelData, err := a.deps.PlayerRepo.GetLevelData(ctx)
	if err != nil {
		return 0, err
	}
	return levelData.Level, nil
}

// GetExp 获取当前经验
func (a *LevelSystemAdapter) GetExp(ctx context.Context) (int64, error) {
	levelData, err := a.deps.PlayerRepo.GetLevelData(ctx)
	if err != nil {
		return 0, err
	}
	return levelData.Exp, nil
}

// GetLevelData 获取等级数据
func (a *LevelSystemAdapter) GetLevelData(ctx context.Context) (*protocol.SiLevelData, error) {
	levelData, err := a.deps.PlayerRepo.GetLevelData(ctx)
	if err != nil {
		return nil, err
	}
	return levelData, nil
}

// CalculateAttrs 计算等级系统的属性（实现属性计算器接口）
// 注意：根据 Clean Architecture 原则，属性计算逻辑应该下沉到专门的属性服务/UseCase。
// 当前实现保留在 Level 系统中是因为需要访问等级数据和配置，后续可考虑：
// 1. 创建独立的属性计算服务，Level 只提供等级数据接口
// 2. 将属性计算逻辑下沉到 UseCase 层，LevelSystemAdapter 只负责数据访问
func (a *LevelSystemAdapter) CalculateAttrs(ctx context.Context) []*protocol.AttrSt {
	levelData, err := a.deps.PlayerRepo.GetLevelData(ctx)
	if err != nil {
		return nil
	}
	// 从配置表获取等级属性（通过 ConfigManager 接口）
	levelAttrs := a.deps.ConfigManager.GetLevelAttrs(levelData.Level)
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

// CalculateAddRate 计算等级加成属性（实现属性加成计算器接口）
// 注意：与 CalculateAttrs 类似，属性加成计算逻辑也应该下沉到专门的属性服务。
// 当前实现保留在 Level 系统中是为了方便访问等级数据和配置，后续可考虑重构。
func (a *LevelSystemAdapter) CalculateAddRate(ctx context.Context, _ *icalc.FightAttrCalc) []*protocol.AttrSt {
	levelSys := GetLevelSys(ctx)
	if levelSys == nil {
		return nil
	}
	level, err := levelSys.GetLevel(ctx)
	if err != nil || level == 0 {
		return nil
	}
	cfg := a.deps.ConfigManager.GetAttrAddRateConfig()
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

// RegisterSystemFactory 注册等级系统工厂（由 register.RegisterAll 调用）
func RegisterSystemFactory(rt *runtime.Runtime) {
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysLevel), func() iface.ISystem {
		return NewLevelSystemAdapter(rt)
	})
}

// 注册属性计算器（保留在 init 中，因为不依赖 Runtime）
func init() {
	// 注册属性计算器
	attrcalc.Register(uint32(protocol.SaAttrSys_SaLevel), func(ctx context.Context) attrcalc.Calculator {
		return GetLevelSys(ctx)
	})

	// 注册属性加成计算器
	attrcalc.RegisterAddRate(uint32(protocol.SaAttrSys_SaLevel), func(ctx context.Context) attrcalc.AddRateCalculator {
		return GetLevelSys(ctx)
	})
}
