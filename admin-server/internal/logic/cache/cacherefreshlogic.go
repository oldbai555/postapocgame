package cache

import (
	"context"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CacheRefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCacheRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CacheRefreshLogic {
	return &CacheRefreshLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CacheRefreshLogic) CacheRefresh() (resp *types.CacheRefreshResp, err error) {
	// 清理所有缓存（字典和配置缓存）
	if err := l.svcCtx.Repository.ClearAllCacheDBs(l.ctx); err != nil {
		logx.Errorf("清理缓存失败: %v", err)
		return nil, err
	}

	logx.Infof("缓存刷新成功")
	return &types.CacheRefreshResp{
		Message: "缓存刷新成功",
	}, nil
}
