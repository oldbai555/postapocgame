// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package audit_log

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuditLogDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuditLogDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditLogDetailLogic {
	return &AuditLogDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuditLogDetailLogic) AuditLogDetail(id uint64) (resp *types.AuditLogDetailResp, err error) {
	if id == 0 {
		return nil, errs.New(errs.CodeBadRequest, "审计日志ID不能为空")
	}

	auditLogRepo := repository.NewAuditLogRepository(l.svcCtx.Repository)
	log, err := auditLogRepo.FindByID(l.ctx, id)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询审计日志详情失败", err)
	}
	if log == nil {
		return nil, errs.New(errs.CodeNotFound, "审计日志不存在")
	}

	auditDetail := ""
	if log.AuditDetail.Valid {
		auditDetail = log.AuditDetail.String
	}

	return &types.AuditLogDetailResp{
		AuditLogItem: types.AuditLogItem{
			Id:          log.Id,
			UserId:      log.UserId,
			Username:    log.Username,
			AuditType:   log.AuditType,
			AuditObject: log.AuditObject,
			AuditDetail: auditDetail,
			IpAddress:   log.IpAddress,
			UserAgent:   log.UserAgent,
			CreatedAt:   log.CreatedAt,
		},
	}, nil
}
