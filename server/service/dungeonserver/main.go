package main

import (
	"context"
	"os"
	"os/signal"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/dungeonserver/internel/devent"
	"postapocgame/server/service/dungeonserver/internel/dshare"
	"postapocgame/server/service/dungeonserver/internel/dungeonactor"
	"postapocgame/server/service/dungeonserver/internel/engine"
	"postapocgame/server/service/dungeonserver/internel/fbmgr"
	"postapocgame/server/service/dungeonserver/internel/gameserverlink"
	"syscall"
	"time"
)

func main() {
	log.InitLogger(log.WithAppName("dungeonserver"), log.WithScreen(true), log.WithPath(tool.GetCurDir()+"log"), log.WithLevel(log.DebugLevel))

	configPath := tool.GetCurDir() + "config"
	if err := jsonconf.GetConfigManager().Init(configPath); err != nil {
		log.Fatalf("init config manager failed: %v", err)
	}

	// 初始化错误码映射
	protocol.InitErrorCodes()
	serverConfig, err := engine.LoadServerConfig("")
	if err != nil {
		log.Fatalf("err:%v", err)
	}

	// 设置DungeonServer类型(用于协议注册)
	gameserverlink.SetDungeonSrvType(serverConfig.SrvType)

	dshare.Codec = network.DefaultCodec()

	// 创建
	ds := engine.NewDungeonServer(serverConfig)

	dungeonActor := dungeonactor.NewDungeonActor(actor.ModeSingle)

	fbMgr := fbmgr.GetFuBenMgr()
	err = fbMgr.CreateDefaultFuBen()
	if err != nil {
		log.Fatalf("err:%v", err)
	}

	// 启动
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := dungeonActor.Start(ctx); err != nil {
		log.Fatalf("Start dungeon actor failed: %v", err)
	}

	if err := ds.Start(ctx); err != nil {
		log.Fatalf("Start DungeonServer failed: %v", err)
	}

	devent.Publish(ctx, event.NewEvent(devent.OnSrvStart))

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Infof("Shutting down DungeonServer...")

	// 注销协议
	gameserverlink.UnregisterProtocols(context.Background())

	// 停止DungeonServer
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := dungeonActor.Stop(shutdownCtx); err != nil {
		log.Errorf("Stop dungeonActor failed: %v", err)
		os.Exit(1)
	}

	if err := ds.Stop(shutdownCtx); err != nil {
		log.Errorf("Stop DungeonServer failed: %v", err)
		os.Exit(1)
	}

	log.Infof("DungeonServer shutdown complete")
}
