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

type LoginLogDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogDetailLogic {
	return &LoginLogDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogDetailLogic) LoginLogDetail(req *types.LoginLogDetailReq) (resp *types.LoginLogDetailResp, err error) {
	if req == nil || req.Id == 0 {
		return nil, errs.New(errs.CodeBadRequest, "登录日志ID不能为空")
	}

	loginLogRepo := repository.NewLoginLogRepository(l.svcCtx.Repository)
	log, err := loginLogRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询登录日志详情失败", err)
	}

	return &types.LoginLogDetailResp{
		LoginLogItem: types.LoginLogItem{
			Id:        log.Id,
			UserId:    log.UserId,
			Username:  log.Username,
			IpAddress: log.IpAddress,
			Location:  log.Location,
			Browser:   log.Browser,
			Os:        log.Os,
			UserAgent: log.UserAgent,
			Status:    int(log.Status),
			Message:   log.Message,
			LoginAt:   log.LoginAt,
			LogoutAt:  log.LogoutAt,
			CreatedAt: log.CreatedAt,
		},
	}, nil
}
