// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type UserCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserCreateLogic {
	return &UserCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserCreateLogic) UserCreate(req *types.UserCreateReq) error {
	if req == nil || req.Username == "" || req.Password == "" {
		return errs.New(errs.CodeBadRequest, "用户名和密码不能为空")
	}

	userRepo := repository.NewUserRepository(l.svcCtx.Repository)
	// 检查用户名是否已存在
	_, err := userRepo.FindByUsername(l.ctx, req.Username)
	if err == nil {
		return errs.New(errs.CodeBadRequest, "用户名已存在")
	}

	// 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "密码加密失败", err)
	}

	user := model.AdminUser{
		Username:     req.Username,
		PasswordHash: string(hash),
		DepartmentId: req.DepartmentId,
		Status:       req.Status,
	}

	if err := userRepo.Create(l.ctx, &user); err != nil {
		return errs.Wrap(errs.CodeInternalError, "创建用户失败", err)
	}
	return nil
}
