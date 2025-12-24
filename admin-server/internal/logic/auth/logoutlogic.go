// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"
	"time"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout(req *types.LogoutReq) error {
	if req == nil {
		return errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	blackRepo := repository.NewTokenBlacklistRepository(l.svcCtx.Repository)

	// 访问令牌加入黑名单
	if req.AccessToken != "" {
		if err := blackRepo.Blacklist(l.ctx, req.AccessToken, time.Duration(l.svcCtx.Config.JWT.AccessExpire)*time.Second); err != nil {
			return errs.Wrap(errs.CodeInternalError, "加入访问令牌黑名单失败", err)
		}
	}

	// 刷新令牌加入黑名单
	if req.RefreshToken != "" {
		if err := blackRepo.Blacklist(l.ctx, req.RefreshToken, time.Duration(l.svcCtx.Config.JWT.RefreshExpire)*time.Second); err != nil {
			return errs.Wrap(errs.CodeInternalError, "加入刷新令牌黑名单失败", err)
		}
	}

	return nil
}
