// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	"postapocgame/admin-server/pkg/initdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserDeleteLogic {
	return &UserDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserDeleteLogic) UserDelete(req *types.UserDeleteReq) error {
	if req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "用户ID不能为空")
	}
	// 保护初始化数据：不允许删除超级管理员用户（id=1）
	if initdata.IsInitUserID(req.Id) {
		return errs.New(errs.CodeBadRequest, "初始化数据不可删除")
	}

	userRepo := repository.NewUserRepository(l.svcCtx.Repository)
	if err := userRepo.DeleteByID(l.ctx, req.Id); err != nil {
		return errs.Wrap(errs.CodeInternalError, "删除用户失败", err)
	}
	return nil
}
