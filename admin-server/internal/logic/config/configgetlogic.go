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

	// 尝试从缓存获取配置
	cache := l.svcCtx.Repository.BusinessCache
	var cachedValue string
	err = cache.GetConfigKey(l.ctx, req.Key, &cachedValue)
	if err == nil {
		// 缓存命中，直接返回
		return &types.ConfigGetResp{
			Value: cachedValue,
		}, nil
	}

	// 缓存未命中，从数据库查询
	configRepo := repository.NewConfigRepository(l.svcCtx.Repository)
	config, err := configRepo.FindByKey(l.ctx, req.Key)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询配置失败", err)
	}

	value := ""
	if config.Value.Valid {
		value = config.Value.String
	}

	resp = &types.ConfigGetResp{
		Value: value,
	}

	// 写入缓存（异步，不阻塞返回）
	go func() {
		if err := cache.SetConfigKey(context.Background(), req.Key, value); err != nil {
			l.Errorf("设置配置缓存失败: key=%s, error=%v", req.Key, err)
		}
	}()

	return resp, nil
}
