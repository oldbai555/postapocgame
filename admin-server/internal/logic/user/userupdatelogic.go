// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type UserUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserUpdateLogic {
	return &UserUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserUpdateLogic) UserUpdate(req *types.UserUpdateReq) error {
	if req == nil || req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "用户ID不能为空")
	}

	userRepo := repository.NewUserRepository(l.svcCtx.Repository)
	user, err := userRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询用户失败", err)
	}

	if req.Username != "" {
		// 检查新用户名是否已被其他用户使用
		existing, err := userRepo.FindByUsername(l.ctx, req.Username)
		if err == nil && existing.Id != req.Id {
			return errs.New(errs.CodeBadRequest, "用户名已被使用")
		}
		user.Username = req.Username
	}

	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return errs.Wrap(errs.CodeInternalError, "密码加密失败", err)
		}
		user.PasswordHash = string(hash)
	}

	if req.DepartmentId != 0 {
		user.DepartmentId = req.DepartmentId
	}

	// 更新昵称、头像和个性签名
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Signature != "" {
		user.Signature = req.Signature
	}

	// Status 字段：0 是有效值（禁用），需要特殊处理
	if req.Status == 0 || req.Status == 1 {
		user.Status = req.Status
	}

	if err := userRepo.Update(l.ctx, user); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新用户失败", err)
	}
	return nil
}
