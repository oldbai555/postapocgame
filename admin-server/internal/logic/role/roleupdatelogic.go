// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"context"
	"database/sql"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleUpdateLogic {
	return &RoleUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RoleUpdateLogic) RoleUpdate(req *types.RoleUpdateReq) error {
	if req == nil || req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "角色ID不能为空")
	}

	roleRepo := repository.NewRoleRepository(l.svcCtx.Repository)
	role, err := roleRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询角色失败", err)
	}

	role.Name = req.Name
	if req.Description != "" {
		role.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.Status != 0 {
		role.Status = req.Status
	}

	if err := roleRepo.Update(l.ctx, role); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新角色失败", err)
	}
	return nil
}
