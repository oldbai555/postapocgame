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

type ConfigGetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfigGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigGetLogic {
	return &ConfigGetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfigGetLogic) ConfigGet(req *types.ConfigGetReq) (resp *types.ConfigGetResp, err error) {
	if req == nil || req.Key == "" {
		return nil, errs.New(errs.CodeBadRequest, "配置键不能为空")
	}

	configRepo := repository.NewConfigRepository(l.svcCtx.Repository)
	config, err := configRepo.FindByKey(l.ctx, req.Key)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询配置失败", err)
	}

	value := ""
	if config.Value.Valid {
		value = config.Value.String
	}

	return &types.ConfigGetResp{
		Value: value,
	}, nil
}
