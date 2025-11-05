package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"postapocgame/server/cmd/gateway/internel"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/routine"
	"syscall"
	"time"
)

// SimpleAuthenticator 简单的认证器示例
type SimpleAuthenticator struct{}

func (a *SimpleAuthenticator) Authenticate(ctx context.Context, token string) (string, error) {
	// TODO: 实现真实的认证逻辑
	// 这里简单返回token作为userID
	if token == "" {
		return "", fmt.Errorf("empty token")
	}
	return "user_" + token, nil
}

func main() {
	log.InitLogger(log.WithAppName("gateway"))
	defer func() {
		log.Flush()
	}()

	// 创建配置
	config, err := internel.LoadGatewayConf("")
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	// 创建认证器
	authenticator := &SimpleAuthenticator{}

	// 创建网关
	gw, err := internel.NewGateway(config, authenticator)
	if err != nil {
		log.Fatalf("create gateway failed: %v", err)
	}

	// 启动网关
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := gw.Start(ctx); err != nil {
		log.Fatalf("start gateway failed: %v", err)
	}

	// 打印统计信息
	routine.GoV2(func() error {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				stats := gw.GetStats()
				log.Infof("Gateway Stats: %+v", stats)
			}
		}
	})

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Infof("shutting down gateway...")

	// 停止网关
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := gw.Stop(shutdownCtx); err != nil {
		log.Errorf("stop gateway failed: %v\n", err)
		os.Exit(1)
	}

	log.Infof("gateway shutdown complete")
}
