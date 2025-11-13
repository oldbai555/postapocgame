package main

import (
	"context"
	"os"
	"os/signal"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/config"
	"postapocgame/server/service/gameserver/internel/dungeonserverlink"
	"postapocgame/server/service/gameserver/internel/engine"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/playeractor"
	"syscall"
	"time"
)

func main() {
	log.InitLogger(log.WithAppName("gameserver"), log.WithScreen(true), log.WithPath(tool.GetCurDir()+"log"))

	// 初始化错误码映射
	protocol.InitErrorCodes()
	serverConfig, err := config.LoadServerConfig("")
	if err != nil {
		log.Fatalf("err:%v", err)
	}

	gshare.SetPlatformId(serverConfig.PlatformID)
	gshare.SetSrvId(serverConfig.SrvId)

	// 创建GameServer
	gs := engine.NewGameServer(serverConfig)

	// 玩家消息处理
	playerRoleActor := playeractor.NewPlayerRoleActor(actor.ModePerKey)
	err = playerRoleActor.Init()
	if err != nil {
		log.Fatalf("err:%v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 连接DungeonServer
	dungeonserverlink.StartDungeonClient(ctx, serverConfig)

	// 初始化协议注册RPC处理器
	dungeonserverlink.InitProtocolRegistration()

	if err := playerRoleActor.Start(ctx); err != nil {
		log.Fatalf("Start playerRoleActor failed: %v", err)
	}

	// 启动GameServer
	if err := gs.Start(ctx); err != nil {
		log.Fatalf("Start GameServer failed: %v", err)
	}

	gevent.Publish(context.Background(), event.NewEvent(gevent.OnSrvStart))

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Infof("Shutting down GameServer...")

	dungeonserverlink.Stop()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := playerRoleActor.Stop(shutdownCtx); err != nil {
		log.Errorf("Stop playerRoleActor failed: %v", err)
	}
	if err := gs.Stop(shutdownCtx); err != nil {
		log.Fatalf("Stop GameServer failed: %v", err)
	}
	log.Infof("GameServer shutdown complete")
}
