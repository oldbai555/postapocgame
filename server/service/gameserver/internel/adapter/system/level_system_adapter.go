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
type LevelSystemAdapter struct {
	*BaseSystemAdapter
	addExpUseCase  *level.AddExpUseCase
	levelUpUseCase *level.LevelUpUseCase
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
		levelUpUseCase: levelUpUC,
	}
}

// OnInit 系统初始化
func (a *LevelSystemAdapter) OnInit(ctx context.Context) {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("level sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果level_data不存在，则初始化
	if binaryData.LevelData == nil {
		binaryData.LevelData = &protocol.SiLevelData{
			Level: 1,
			Exp:   0,
		}
	}

	// 确保等级至少为1
	if binaryData.LevelData.Level < 1 {
		binaryData.LevelData.Level = 1
	}
	if binaryData.LevelData.Exp < 0 {
		binaryData.LevelData.Exp = 0
	}

	// 同步经验到货币系统（经验作为货币的一种）
	// 统一以等级系统的经验值为准，同步到货币系统
	if binaryData.MoneyData != nil {
		if binaryData.MoneyData.MoneyMap == nil {
			binaryData.MoneyData.MoneyMap = make(map[uint32]int64)
		}
		expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
		// 统一以等级系统的经验值为准
		binaryData.MoneyData.MoneyMap[expMoneyID] = binaryData.LevelData.Exp
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
