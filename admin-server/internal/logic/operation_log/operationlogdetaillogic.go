// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package operation_log

import (
	"context"
	"fmt"
	"time"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperationLogDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOperationLogDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperationLogDetailLogic {
	return &OperationLogDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OperationLogDetailLogic) OperationLogDetail(idStr string) (resp *types.OperationLogDetailResp, err error) {
	if idStr == "" {
		return nil, errs.New(errs.CodeBadRequest, "操作日志ID不能为空")
	}

	var id uint64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return nil, errs.New(errs.CodeBadRequest, "操作日志ID格式错误")
	}

	operationLogRepo := repository.NewOperationLogRepository(l.svcCtx.Repository)
	log, err := operationLogRepo.FindByID(l.ctx, id)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询操作日志失败", err)
	}

	requestParams := ""
	if log.RequestParams.Valid {
		requestParams = log.RequestParams.String
	}

	item := types.OperationLogItem{
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
	}

	return &types.OperationLogDetailResp{
		OperationLog: item,
	}, nil
}
