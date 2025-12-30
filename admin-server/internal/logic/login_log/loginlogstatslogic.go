// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package login_log

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogStatsLogic {
	return &LoginLogStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogStatsLogic) LoginLogStats() (resp *types.LoginLogStatsResp, err error) {
	loginLogRepo := repository.NewLoginLogRepository(l.svcCtx.Repository)

	// 总登录次数（查询所有状态，使用一个大的查询）
	var totalCount int64
	// 由于 CountByStatus 不支持 -1，我们分别查询成功和失败，然后相加
	successCount, err := loginLogRepo.CountByStatus(l.ctx, 1)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询成功次数失败", err)
	}
	failureCount, err := loginLogRepo.CountByStatus(l.ctx, 0)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询失败次数失败", err)
	}
	totalCount = successCount + failureCount

	// 今日登录次数
	todayCount, err := loginLogRepo.CountToday(l.ctx)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询今日登录次数失败", err)
	}

	// 今日成功次数
	todaySuccess, err := loginLogRepo.CountTodayByStatus(l.ctx, 1)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询今日成功次数失败", err)
	}

	// 今日失败次数
	todayFailure, err := loginLogRepo.CountTodayByStatus(l.ctx, 0)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询今日失败次数失败", err)
	}

	// 当前在线用户数（从 WebSocket Hub 获取）
	// 注意：ChatOnlineUser 表已移除，在线用户数从 WebSocket Hub 获取
	onlineUserCount := int64(0)
	if l.svcCtx.ChatHub != nil {
		onlineUserIDs := l.svcCtx.ChatHub.GetOnlineUsers()
		onlineUserCount = int64(len(onlineUserIDs))
	}

	return &types.LoginLogStatsResp{
		TotalCount:      totalCount,
		SuccessCount:    successCount,
		FailureCount:    failureCount,
		TodayCount:      todayCount,
		TodaySuccess:    todaySuccess,
		TodayFailure:    todayFailure,
		OnlineUserCount: onlineUserCount,
	}, nil
}
