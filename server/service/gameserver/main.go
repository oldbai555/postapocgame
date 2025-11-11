package main

import (
	"context"
	"os"
	"os/signal"
	"postapocgame/server/internal/actor"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/config"
	"postapocgame/server/service/gameserver/internel/dungeonserverlink"
	"postapocgame/server/service/gameserver/internel/engine"
	"postapocgame/server/service/gameserver/internel/playeractor"
	"postapocgame/server/service/gameserver/internel/pubilcactor"
	"syscall"
	"time"
)

func main() {
	log.InitLogger(log.WithAppName("gameserver"), log.WithScreen(true), log.WithPath(tool.GetCurDir()+"log"))
	serverConfig, err := config.LoadServerConfig("")
	if err != nil {
		log.Fatalf("err:%v", err)
	}

	// 创建GameServer
	gs := engine.NewGameServer(serverConfig)

	// 玩家消息处理
	playerRoleActor := playeractor.NewActorSystem(serverConfig.ActorMode)
	playerRoleActor.Init()

	// 公共消息处理
	publicActor := pubilcactor.NewActorSystem()
	publicActor.Init()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 连接DungeonServer
	dungeonserverlink.StartDungeonClient(ctx, serverConfig)

	if err := playerRoleActor.Start(ctx); err != nil {
		log.Fatalf("Start playerRoleActor failed: %v", err)
	}

	if err := publicActor.Start(ctx); err != nil {
		log.Fatalf("Start publicActor failed: %v", err)
	}

	// 监控
	actor.GetActorMonitor().Start(ctx, time.Hour)

	// 启动GameServer
	if err := gs.Start(ctx); err != nil {
		log.Fatalf("Start GameServer failed: %v", err)
	}

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Infof("Shutting down GameServer...")
	actor.GetActorMonitor().Stop()

	// 1. 先关闭 DungeonClient（优雅关闭）
	dungeonserverlink.Stop()

	// 2. 停止 Actor 系统
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := playerRoleActor.Stop(shutdownCtx); err != nil {
		log.Errorf("Stop playerRoleActor failed: %v", err)
	}

	if err := publicActor.Stop(shutdownCtx); err != nil {
		log.Errorf("Stop publicActor failed: %v", err)
	}

	// 3. 最后停止 GameServer
	if err := gs.Stop(shutdownCtx); err != nil {
		log.Fatalf("Stop GameServer failed: %v", err)
	}

	log.Infof("GameServer shutdown complete")
}
