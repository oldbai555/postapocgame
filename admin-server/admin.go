// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/zeromicro/go-zero/core/logx"

	"postapocgame/admin-server/internal/config"
	"postapocgame/admin-server/internal/handler"
	"postapocgame/admin-server/internal/middleware"
	"postapocgame/admin-server/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/admin-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	err := logx.SetUp(logx.LogConf{
		Encoding: "plain",
	})
	if err != nil {
		log.Fatalf("Failed to set up logging: %v", err)
	}

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		log.Fatalf("init service context: %v", err)
	}

	// 初始化中间件（避免循环依赖，在外部初始化）
	authMiddleware := middleware.NewAuthMiddleware(ctx)
	permissionMiddleware := middleware.NewPermissionMiddleware(ctx)
	ctx.AuthMiddleware = authMiddleware.Handle
	ctx.PermissionMiddleware = permissionMiddleware.Handle

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
