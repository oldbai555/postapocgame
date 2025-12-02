package system

import (
	"context"
	"postapocgame/server/service/gameserver/internel/core/iface"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	domain "postapocgame/server/service/gameserver/internel/domain/dailyactivity"
)

// DailyActivitySystemAdapter 日常活跃系统适配器
type DailyActivitySystemAdapter struct {
	*BaseSystemAdapter
	data *protocol.SiDailyActivityData
}

// NewDailyActivitySystemAdapter 创建适配器
func NewDailyActivitySystemAdapter() *DailyActivitySystemAdapter {
	return &DailyActivitySystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysDailyActivity)),
	}
}

// OnInit 初始化数据
func (a *DailyActivitySystemAdapter) OnInit(ctx context.Context) {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("daily activity OnInit get role err:%v", err)
		return
	}
	data := domain.EnsureData(playerRole.GetBinaryData())
	if domain.NeedReset(data, servertime.Now()) {
		domain.ResetForNewDay(data, servertime.Now())
	}
	a.data = data
}

// GetData 获取活跃数据
func (a *DailyActivitySystemAdapter) GetData(ctx context.Context) *protocol.SiDailyActivityData {
	if a.data != nil {
		return a.data
	}
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		return nil
	}
	a.data = domain.EnsureData(playerRole.GetBinaryData())
	return a.data
}

// OnRoleLogin 登录时做一次重置检查
func (a *DailyActivitySystemAdapter) OnRoleLogin(ctx context.Context) {
	data := a.GetData(ctx)
	if data == nil {
		return
	}
	now := servertime.Now()
	if domain.NeedReset(data, now) {
		domain.ResetForNewDay(data, now)
	}
}

// 确保实现 ISystem 接口
var _ iface.ISystem = (*DailyActivitySystemAdapter)(nil)

// AddActivePoints 添加活跃点（供 DailyActivityUseCaseAdapter 调用）
func (a *DailyActivitySystemAdapter) AddActivePoints(ctx context.Context, points uint32) error {
	data := a.GetData(ctx)
	if data == nil {
		return nil
	}
	data.TodayPoints += points
	return nil
}
