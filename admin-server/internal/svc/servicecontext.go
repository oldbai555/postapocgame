package svc

import (
	"postapocgame/admin-server/internal/config"
	"postapocgame/admin-server/internal/hub"
	"postapocgame/admin-server/internal/repository"

	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config                 config.Config
	Repository             *repository.Repository
	ChatHub                *hub.ChatHub
	AuthMiddleware         rest.Middleware
	PermissionMiddleware   rest.Middleware
	OperationLogMiddleware rest.Middleware
	RateLimitMiddleware    rest.Middleware
	PerformanceMiddleware  rest.Middleware
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	repo, err := repository.BuildSources(c)
	if err != nil {
		return nil, err
	}

	// 初始化 ChatHub（传入在线用户 Repository）
	onlineUserRepo := repository.NewChatOnlineUserRepository(repo)
	chatHub := hub.NewChatHub(onlineUserRepo)
	go chatHub.Run()

	return &ServiceContext{
		Config:     c,
		Repository: repo,
		ChatHub:    chatHub,
		// AuthMiddleware 和 PermissionMiddleware 需要在外部初始化，避免循环依赖
	}, nil
}
