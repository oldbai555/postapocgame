// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfigListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfigListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigListLogic {
	return &ConfigListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfigListLogic) ConfigList(req *types.ConfigListReq) (resp *types.ConfigListResp, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	configRepo := repository.NewConfigRepository(l.svcCtx.Repository)
	list, total, err := configRepo.FindPage(l.ctx, req.Page, req.PageSize, req.Group, req.Key)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询配置列表失败", err)
	}

	items := make([]types.ConfigItem, 0, len(list))
	for _, c := range list {
		value := ""
		if c.Value.Valid {
			value = c.Value.String
		}
		description := ""
		if c.Description.Valid {
			description = c.Description.String
		}
		items = append(items, types.ConfigItem{
			Id:          c.Id,
			Group:       c.Group,
			Key:         c.Key,
			Value:       value,
			ConfigType:  c.Type,
			Description: description,
			CreatedAt:   c.CreatedAt,
		})
	}

	return &types.ConfigListResp{
		Total: total,
		List:  items,
	}, nil
}
