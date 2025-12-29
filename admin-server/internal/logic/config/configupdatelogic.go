// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"context"
	"database/sql"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfigUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfigUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigUpdateLogic {
	return &ConfigUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfigUpdateLogic) ConfigUpdate(req *types.ConfigUpdateReq) error {
	if req == nil || req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "配置ID不能为空")
	}

	configRepo := repository.NewConfigRepository(l.svcCtx.Repository)
	config, err := configRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询配置失败", err)
	}

	if req.Value != "" {
		config.Value = sql.NullString{String: req.Value, Valid: true}
	}
	if req.Description != "" {
		config.Description = sql.NullString{String: req.Description, Valid: true}
	}

	if err := configRepo.Update(l.ctx, config); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新配置失败", err)
	}

	// 清除配置缓存
	cache := l.svcCtx.Repository.BusinessCache
	go func() {
		if err := cache.DeleteConfigKey(context.Background(), config.Key); err != nil {
			l.Errorf("清除配置缓存失败: key=%s, error=%v", config.Key, err)
		}
	}()

	return nil
}
