package main

import (
	"context"
	"os"
	"os/signal"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gateway/internel/engine"
	"syscall"
	"time"
)

func main() {
	log.InitLogger(log.WithAppName("gateway"), log.WithScreen(true), log.WithPath(tool.GetCurDir()+"log"))

	// 初始化错误码映射
	protocol.InitErrorCodes()
	defer func() {
		log.Flush()
	}()

	// 创建配置
	config, err := engine.LoadGatewayConf("")
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	// 创建网关
	gw, err := engine.NewGatewayServer(config)
	if err != nil {
		log.Fatalf("create gateway failed: %v", err)
	}

	// 启动网关
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := gw.Start(ctx); err != nil {
		log.Fatalf("start gateway failed: %v", err)
	}

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Infof("shutting down gateway...")

	// 停止网关
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := gw.Stop(shutdownCtx); err != nil {
		log.Fatalf("stop gateway failed: %v", err)
	}

	log.Infof("gateway shutdown complete")
}
