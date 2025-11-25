package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"postapocgame/server/example/internal/client"
	"postapocgame/server/example/internal/panel"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
)

func main() {
	log.InitLogger(
		log.WithAppName("example_client_actor"),
		log.WithScreen(true),
		log.WithPath(tool.GetCurDir()+"log"),
	)
	defer log.Flush()

	configPath := tool.GetCurDir() + "config"
	if err := jsonconf.GetConfigManager().Init(configPath); err != nil {
		log.Fatalf("init config manager failed: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	manager := client.NewManager(ctx)
	defer manager.Stop()

	panel := panel.NewAdventurePanel(ctx, manager)
	if err := panel.Run(); err != nil && !errors.Is(err, context.Canceled) {
		log.Errorf("⚠️ 文字冒险面板退出: %v", err)
	}
}
