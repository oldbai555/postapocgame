package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/level"
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
type LevelSystemAdapter struct {
	*BaseSystemAdapter
	addExpUseCase        *level.AddExpUseCase
	levelUpUseCase       *level.LevelUpUseCase
	initLevelDataUseCase *level.InitLevelDataUseCase
}

// NewLevelSystemAdapter 创建等级系统适配器
func NewLevelSystemAdapter() *LevelSystemAdapter {
	container := di.GetContainer()
	levelUpUC := level.NewLevelUpUseCase(
		container.PlayerGateway(),
		container.EventPublisher(),
		container.ConfigGateway(),
	)

	// 注入 AttrUseCase 依赖（通过适配器）
	attrUseCase := NewAttrUseCaseAdapter()
	levelUpUC.SetDependencies(nil, attrUseCase) // moneyUseCase 暂时为 nil，等 MoneySys 完全重构后注入

	return &LevelSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysLevel)),
		addExpUseCase: level.NewAddExpUseCase(
			container.PlayerGateway(),
			container.EventPublisher(),
			container.ConfigGateway(),
		),
		levelUpUseCase:       levelUpUC,
		initLevelDataUseCase: level.NewInitLevelDataUseCase(container.PlayerGateway()),
	}
}

// OnInit 系统初始化
func (a *LevelSystemAdapter) OnInit(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
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
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	return a.addExpUseCase.Execute(ctx, roleID, exp)
}

// GetLevel 获取当前等级
func (a *LevelSystemAdapter) GetLevel(ctx context.Context) (uint32, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return 0, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return 0, err
	}
	if binaryData.LevelData == nil {
		return 1, nil
	}
	return binaryData.LevelData.Level, nil
}

// GetExp 获取当前经验
func (a *LevelSystemAdapter) GetExp(ctx context.Context) (int64, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return 0, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return 0, err
	}
	if binaryData.LevelData == nil {
		return 0, nil
	}
	return binaryData.LevelData.Exp, nil
}

// GetLevelData 获取等级数据
func (a *LevelSystemAdapter) GetLevelData(ctx context.Context) (*protocol.SiLevelData, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if binaryData.LevelData == nil {
		return &protocol.SiLevelData{Level: 1, Exp: 0}, nil
	}
	return binaryData.LevelData, nil
}

// EnsureISystem 确保 LevelSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*LevelSystemAdapter)(nil)
