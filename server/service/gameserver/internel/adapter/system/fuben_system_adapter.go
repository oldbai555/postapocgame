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
type FubenSystemAdapter struct {
	*BaseSystemAdapter
	enterDungeonUseCase *fuben.EnterDungeonUseCase
}

// NewFubenSystemAdapter 创建副本系统适配器
func NewFubenSystemAdapter() *FubenSystemAdapter {
	container := di.GetContainer()
	enterDungeonUC := fuben.NewEnterDungeonUseCase(container.PlayerGateway(), container.ConfigGateway(), container.DungeonServerGateway())

	// 注入依赖
	consumeUseCase := usecaseadapter.NewConsumeUseCaseAdapter()
	enterDungeonUC.SetDependencies(consumeUseCase)

	return &FubenSystemAdapter{
		BaseSystemAdapter:   NewBaseSystemAdapter(uint32(protocol.SystemId_SysFuBen)),
		enterDungeonUseCase: enterDungeonUC,
	}
}

// OnInit 系统初始化
func (a *FubenSystemAdapter) OnInit(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("fuben sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		log.Errorf("fuben sys OnInit get binary data err:%v", err)
		return
	}

	// 如果dungeon_data不存在，则初始化
	if binaryData.DungeonData == nil {
		binaryData.DungeonData = &protocol.SiDungeonData{
			Records: make([]*protocol.DungeonRecord, 0),
		}
	}

	// 如果Records为空，初始化为空切片
	if binaryData.DungeonData.Records == nil {
		binaryData.DungeonData.Records = make([]*protocol.DungeonRecord, 0)
	}

	log.Infof("FuBenSys initialized: RecordCount=%d", len(binaryData.DungeonData.Records))
}

// GetDungeonRecord 获取副本记录（对外接口，供其他系统调用）
func (a *FubenSystemAdapter) GetDungeonRecord(ctx context.Context, dungeonID uint32, difficulty uint32) (*protocol.DungeonRecord, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if binaryData.DungeonData == nil || binaryData.DungeonData.Records == nil {
		return nil, nil
	}
	for _, record := range binaryData.DungeonData.Records {
		if record != nil && record.DungeonId == dungeonID && record.Difficulty == difficulty {
			return record, nil
		}
	}
	return nil, nil
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
