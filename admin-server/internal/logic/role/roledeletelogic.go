// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	"postapocgame/admin-server/pkg/initdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleDeleteLogic {
	return &RoleDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RoleDeleteLogic) RoleDelete(req *types.RoleDeleteReq) error {
	if req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "角色ID不能为空")
	}
	// 保护初始化数据：不允许删除超级管理员角色（id=1）
	if initdata.IsInitRoleID(req.Id) {
		return errs.New(errs.CodeBadRequest, "初始化数据不可删除")
	}

	roleRepo := repository.NewRoleRepository(l.svcCtx.Repository)
	if err := roleRepo.DeleteByID(l.ctx, req.Id); err != nil {
		return errs.Wrap(errs.CodeInternalError, "删除角色失败", err)
	}
	return nil
}
