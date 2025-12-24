// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"context"
	"database/sql"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleCreateLogic {
	return &RoleCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RoleCreateLogic) RoleCreate(req *types.RoleCreateReq) error {
	if req == nil || req.Name == "" || req.Code == "" {
		return errs.New(errs.CodeBadRequest, "角色名称和编码不能为空")
	}

	roleRepo := repository.NewRoleRepository(l.svcCtx.Repository)
	role := model.AdminRole{
		Name:        req.Name,
		Code:        req.Code,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Status:      req.Status,
	}

	if err := roleRepo.Create(l.ctx, &role); err != nil {
		return errs.Wrap(errs.CodeInternalError, "创建角色失败", err)
	}
	return nil
}
