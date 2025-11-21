package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
)

func main() {
	log.InitLogger(log.WithAppName("example_client_actor"), log.WithScreen(true), log.WithPath(tool.GetCurDir()+"log"))
	defer log.Flush()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	clientMgr := NewClientManager(ctx)
	defer clientMgr.Stop()

	panel := NewAdventurePanel(ctx, clientMgr)
	if err := panel.Run(); err != nil && !errors.Is(err, context.Canceled) {
		log.Errorf("⚠️ 文字冒险面板退出: %v", err)
	}
}
