// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type PermissionListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPermissionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PermissionListLogic {
	return &PermissionListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PermissionListLogic) PermissionList(req *types.PermissionListReq) (resp *types.PermissionListResp, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	permissionRepo := repository.NewPermissionRepository(l.svcCtx.Repository)
	list, total, err := permissionRepo.FindPage(l.ctx, req.Page, req.PageSize, req.Name)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询权限列表失败", err)
	}

	items := make([]types.PermissionItem, 0, len(list))
	for _, p := range list {
		description := ""
		if p.Description.Valid {
			description = p.Description.String
		}
		items = append(items, types.PermissionItem{
			Id:          p.Id,
			Name:        p.Name,
			Code:        p.Code,
			Description: description,
		})
	}

	return &types.PermissionListResp{
		Total: total,
		List:  items,
	}, nil
}
