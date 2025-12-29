// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package performance_log

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PerformanceLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPerformanceLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PerformanceLogListLogic {
	return &PerformanceLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PerformanceLogListLogic) PerformanceLogList(req *types.PerformanceLogListReq) (resp *types.PerformanceLogListResp, err error) {
	// 参数兜底处理
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 仓库查询
	perfRepo := repository.NewPerformanceLogRepository(l.svcCtx.Repository)
	logs, total, err := perfRepo.FindPage(
		l.ctx,
		page,
		pageSize,
		req.Method,
		req.Path,
		req.IsSlow,
		req.StatusCode,
		req.StartTime,
		req.EndTime,
	)
	if err != nil {
		return nil, err
	}

	// 转为响应结构
	items := make([]types.PerformanceLogItem, 0, len(logs))
	for _, lg := range logs {
		items = append(items, types.PerformanceLogItem{
			Id:            lg.Id,
			UserId:        lg.UserId,
			Username:      lg.Username,
			Method:        lg.Method,
			Path:          lg.Path,
			StatusCode:    lg.StatusCode,
			Duration:      lg.Duration,
			IsSlow:        lg.IsSlow,
			SlowThreshold: lg.SlowThreshold,
			IpAddress:     lg.IpAddress,
			UserAgent:     lg.UserAgent,
			ErrorMsg:      lg.ErrorMsg,
			CreatedAt:     lg.CreatedAt,
		})
	}

	return &types.PerformanceLogListResp{
		Total: total,
		List:  items,
	}, nil
}
