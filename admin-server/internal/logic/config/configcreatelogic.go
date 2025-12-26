// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

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

type ConfigCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfigCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigCreateLogic {
	return &ConfigCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfigCreateLogic) ConfigCreate(req *types.ConfigCreateReq) error {
	if req == nil || req.Group == "" || req.Key == "" {
		return errs.New(errs.CodeBadRequest, "配置分组和键不能为空")
	}

	configRepo := repository.NewConfigRepository(l.svcCtx.Repository)
	// 检查配置键是否已存在
	_, err := configRepo.FindByKey(l.ctx, req.Key)
	if err == nil {
		return errs.New(errs.CodeBadRequest, "配置键已存在")
	}

	configType := req.ConfigType
	if configType == "" {
		configType = "string"
	}

	config := model.AdminConfig{
		Group:       req.Group,
		Key:         req.Key,
		Value:       sql.NullString{String: req.Value, Valid: req.Value != ""},
		Type:        configType,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
	}

	if err := configRepo.Create(l.ctx, &config); err != nil {
		return errs.Wrap(errs.CodeInternalError, "创建配置失败", err)
	}
	return nil
}
