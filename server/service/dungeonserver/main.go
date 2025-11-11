package main

import (
	"context"
	"os"
	"os/signal"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/dungeonserver/internel/config"
	"postapocgame/server/service/dungeonserver/internel/devent"
	"postapocgame/server/service/dungeonserver/internel/dungeonactor"
	"postapocgame/server/service/dungeonserver/internel/engine"
	"postapocgame/server/service/dungeonserver/internel/fbmgr"
	"syscall"
	"time"
)

func main() {
	log.InitLogger(log.WithAppName("dungeonserver"), log.WithScreen(true), log.WithPath(tool.GetCurDir()+"log"))
	serverConfig, err := config.LoadServerConfig("")
	if err != nil {
		log.Fatalf("err:%v", err)
	}

	// 创建
	ds := engine.NewDungeonServer(serverConfig)

	actorSystem := dungeonactor.NewActorSystem()
	actorSystem.Init()

	fbMgr := fbmgr.GetFuBenMgr()
	err = fbMgr.CreateDefaultFuBen()
	if err != nil {
		log.Fatalf("err:%v", err)
	}

	// 启动
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := actorSystem.Start(ctx); err != nil {
		log.Fatalf("Start dungeon actor failed: %v", err)
	}

	fbMgr.StartCleanupRoutine(ctx)

	if err := ds.Start(ctx); err != nil {
		log.Fatalf("Start DungeonServer failed: %v", err)
	}

	err = devent.Publish(ctx, event.NewEvent(1, "main"))
	if err != nil {
		log.Fatalf("Start DungeonServer failed: %v", err)
		return
	}

	// 监控
	actor.GetActorMonitor().Start(ctx, time.Hour)

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Infof("Shutting down DungeonServer...")
	actor.GetActorMonitor().Stop()

	// 停止DungeonServer
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := ds.Stop(shutdownCtx); err != nil {
		log.Errorf("Stop DungeonServer failed: %v", err)
		os.Exit(1)
	}

	log.Infof("DungeonServer shutdown complete")
}
