package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/usecaseadapter"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/fuben"
)

// FubenSystemAdapter 副本系统适配器
//
// 生命周期职责：
// - OnInit: 调用 InitDungeonDataUseCase 初始化副本数据（副本记录容器结构）
// - 其他生命周期: 暂未使用
//
// 业务逻辑：所有业务逻辑（进入副本、副本结算、记录查找）均在 UseCase 层实现
// 外部交互：通过 DungeonServerGateway 进行副本进入/结算 RPC 调用
//
// ⚠️ 防退化机制：禁止在 SystemAdapter 中编写业务规则逻辑，只允许调用 UseCase 与管理生命周期
type FubenSystemAdapter struct {
	*BaseSystemAdapter
	enterDungeonUseCase     *fuben.EnterDungeonUseCase
	initDungeonDataUseCase  *fuben.InitDungeonDataUseCase
	getDungeonRecordUseCase *fuben.GetDungeonRecordUseCase
}

// NewFubenSystemAdapter 创建副本系统适配器
func NewFubenSystemAdapter() *FubenSystemAdapter {
	container := di.GetContainer()
	enterDungeonUC := fuben.NewEnterDungeonUseCase(container.PlayerGateway(), container.ConfigGateway(), container.DungeonServerGateway())
	initDungeonDataUC := fuben.NewInitDungeonDataUseCase(container.PlayerGateway())
	getDungeonRecordUC := fuben.NewGetDungeonRecordUseCase(container.PlayerGateway())

	// 注入依赖
	consumeUseCase := usecaseadapter.NewConsumeUseCaseAdapter()
	enterDungeonUC.SetDependencies(consumeUseCase)

	return &FubenSystemAdapter{
		BaseSystemAdapter:       NewBaseSystemAdapter(uint32(protocol.SystemId_SysFuBen)),
		enterDungeonUseCase:     enterDungeonUC,
		initDungeonDataUseCase:  initDungeonDataUC,
		getDungeonRecordUseCase: getDungeonRecordUC,
	}
}

// OnInit 系统初始化
func (a *FubenSystemAdapter) OnInit(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("fuben sys OnInit get role err:%v", err)
		return
	}
	// 初始化副本数据（包括副本记录容器结构等业务逻辑）
	if err := a.initDungeonDataUseCase.Execute(ctx, roleID); err != nil {
		log.Errorf("fuben sys OnInit init dungeon data err:%v", err)
		return
	}
	// 获取记录数量用于日志（可选）
	binaryData, _ := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	recordCount := 0
	if binaryData != nil && binaryData.DungeonData != nil {
		recordCount = len(binaryData.DungeonData.Records)
	}
	log.Infof("FuBenSys initialized: RecordCount=%d", recordCount)
}

// GetDungeonRecord 获取副本记录（对外接口，供其他系统调用）
func (a *FubenSystemAdapter) GetDungeonRecord(ctx context.Context, dungeonID uint32, difficulty uint32) (*protocol.DungeonRecord, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// 使用 UseCase 查找副本记录（纯业务逻辑已下沉）
	return a.getDungeonRecordUseCase.Execute(ctx, roleID, dungeonID, difficulty)
}

// GetDungeonData 获取副本数据（用于协议）
func (a *FubenSystemAdapter) GetDungeonData(ctx context.Context) (*protocol.SiDungeonData, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if binaryData.DungeonData == nil {
		return &protocol.SiDungeonData{
			Records: make([]*protocol.DungeonRecord, 0),
		}, nil
	}
	return binaryData.DungeonData, nil
}

// EnsureISystem 确保 FubenSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*FubenSystemAdapter)(nil)
