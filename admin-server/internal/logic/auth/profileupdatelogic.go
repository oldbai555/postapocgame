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
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProfileUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProfileUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileUpdateLogic {
	return &ProfileUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProfileUpdateLogic) ProfileUpdate(req *types.ProfileUpdateReq) error {
	if req == nil {
		return errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	// 获取当前用户
	user, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	// 获取用户信息
	userRepo := repository.NewUserRepository(l.svcCtx.Repository)
	userInfo, err := userRepo.FindByID(l.ctx, user.UserID)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "获取用户信息失败", err)
	}

	// 更新头像和个性签名
	if req.Avatar != "" {
		userInfo.Avatar = req.Avatar
	}
	if req.Signature != "" {
		userInfo.Signature = req.Signature
	}

	userInfo.UpdatedAt = time.Now().Unix()

	// 保存更新
	err = userRepo.Update(l.ctx, userInfo)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新个人信息失败", err)
	}

	return nil
}
