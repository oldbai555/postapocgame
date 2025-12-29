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
	"golang.org/x/crypto/bcrypt"
)

type PasswordChangeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasswordChangeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PasswordChangeLogic {
	return &PasswordChangeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PasswordChangeLogic) PasswordChange(req *types.PasswordChangeReq) error {
	if req == nil {
		return errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	// 验证参数
	if req.OldPassword == "" {
		return errs.New(errs.CodeBadRequest, "原密码不能为空")
	}
	if req.NewPassword == "" {
		return errs.New(errs.CodeBadRequest, "新密码不能为空")
	}
	if len(req.NewPassword) < 6 {
		return errs.New(errs.CodeBadRequest, "新密码长度不能少于6位")
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

	// 验证原密码
	err = bcrypt.CompareHashAndPassword([]byte(userInfo.PasswordHash), []byte(req.OldPassword))
	if err != nil {
		return errs.New(errs.CodeBadRequest, "原密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "加密密码失败", err)
	}

	// 更新密码
	userInfo.PasswordHash = string(hashedPassword)
	userInfo.UpdatedAt = time.Now().Unix()

	err = userRepo.Update(l.ctx, userInfo)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新密码失败", err)
	}

	return nil
}
