package svc

import (
	"github.com/zeromicro/go-zero/rest"
	"postapocgame/admin-server/internal/config"
	"postapocgame/admin-server/internal/repository"
)

type ServiceContext struct {
	Config               config.Config
	Repository           *repository.Repository
	AuthMiddleware       rest.Middleware
	PermissionMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	repo, err := repository.BuildSources(c)
	if err != nil {
		return nil, err
	}

	return &ServiceContext{
		Config:     c,
		Repository: repo,
		// AuthMiddleware 和 PermissionMiddleware 需要在外部初始化，避免循环依赖
	}, nil
}
