// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package operation_log

import (
	"context"
	"time"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperationLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOperationLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperationLogListLogic {
	return &OperationLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OperationLogListLogic) OperationLogList(req *types.OperationLogListReq) (resp *types.OperationLogListResp, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	operationLogRepo := repository.NewOperationLogRepository(l.svcCtx.Repository)
	list, total, err := operationLogRepo.FindPage(
		l.ctx,
		req.Page,
		req.PageSize,
		req.UserId,
		req.Username,
		req.OperationType,
		req.OperationObject,
		req.Method,
		req.StartTime,
		req.EndTime,
	)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询操作日志列表失败", err)
	}

	items := make([]types.OperationLogItem, 0, len(list))
	for _, log := range list {
		requestParams := ""
		if log.RequestParams.Valid {
			requestParams = log.RequestParams.String
		}

		items = append(items, types.OperationLogItem{
			Id:              log.Id,
			UserId:          log.UserId,
			Username:        log.Username,
			OperationType:   log.OperationType,
			OperationObject: log.OperationObject,
			Method:          log.Method,
			Path:            log.Path,
			RequestParams:   requestParams,
			ResponseCode:    int(log.ResponseCode),
			ResponseMsg:     log.ResponseMsg,
			IpAddress:       log.IpAddress,
			UserAgent:       log.UserAgent,
			Duration:        int(log.Duration),
			CreatedAt:       time.Unix(log.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		})
	}

	return &types.OperationLogListResp{
		Total: total,
		List:  items,
	}, nil
}
