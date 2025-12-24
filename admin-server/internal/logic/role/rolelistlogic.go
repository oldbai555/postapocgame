// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleListLogic {
	return &RoleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RoleListLogic) RoleList(req *types.RoleListReq) (resp *types.RoleListResp, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	roleRepo := repository.NewRoleRepository(l.svcCtx.Repository)
	list, total, err := roleRepo.FindPage(l.ctx, req.Page, req.PageSize, req.Name)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询角色列表失败", err)
	}

	items := make([]types.RoleItem, 0, len(list))
	for _, r := range list {
		description := ""
		if r.Description.Valid {
			description = r.Description.String
		}
		items = append(items, types.RoleItem{
			Id:          r.Id,
			Name:        r.Name,
			Code:        r.Code,
			Description: description,
			Status:      r.Status,
		})
	}

	return &types.RoleListResp{
		Total: total,
		List:  items,
	}, nil
}
