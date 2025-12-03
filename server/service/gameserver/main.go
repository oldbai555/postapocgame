package main

import (
	"context"
	"os"
	"os/signal"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor"
	"postapocgame/server/service/gameserver/internel/app/engine"
	"postapocgame/server/service/gameserver/internel/app/manager"
	"postapocgame/server/service/gameserver/internel/app/playeractor"
	"postapocgame/server/service/gameserver/internel/app/publicactor"
	"postapocgame/server/service/gameserver/internel/core/gshare"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
	"syscall"
	"time"
)

func main() {
	log.InitLogger(log.WithAppName("gameserver"), log.WithScreen(true), log.WithPath(tool.GetCurDir()+"log"), log.WithLevel(log.DebugLevel))

	configPath := tool.GetCurDir() + "config"
	if err := jsonconf.GetConfigManager().Init(configPath); err != nil {
		log.Fatalf("init config manager failed: %v", err)
	}

	// 初始化数据库
	dbPath := tool.GetCurDir() + "postapocgame.db"
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("数据表自动迁移失败: %v", err)
	}
	log.Infof("数据库初始化成功: %s", dbPath)

	// 初始化错误码映射
	protocol.InitErrorCodes()
	serverConfig, err := engine.LoadServerConfig("")
	if err != nil {
		log.Fatalf("err:%v", err)
	}

	gshare.SetPlatformId(serverConfig.PlatformID)
	gshare.SetSrvId(serverConfig.SrvId)

	serverInfo, err := database.EnsureServerInfo(serverConfig.PlatformID, serverConfig.SrvId)
	if err != nil {
		log.Fatalf("ensure server info failed: %v", err)
	}
	gshare.SetOpenSrvTime(serverInfo.ServerOpenTimeAt)

	// 创建GameServer
	gs := engine.NewGameServer(serverConfig)

	// 玩家消息处理
	playerRoleActor := playeractor.NewPlayerRoleActor(actor.ModePerKey)
	err = playerRoleActor.Init()
	if err != nil {
		log.Fatalf("err:%v", err)
	}

	// 公共Actor（社交经济系统）
	publicActor := publicactor.NewPublicActor()
	err = publicActor.Init()
	if err != nil {
		log.Fatalf("init publicActor failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 副本 / 战斗 DungeonActor（单 Actor，常驻运行）
	dActor := dungeonactor.NewDungeonActor(actor.ModeSingle)

	if err := playerRoleActor.Start(ctx); err != nil {
		log.Fatalf("Start playerRoleActor failed: %v", err)
	}

	if err := publicActor.Start(ctx); err != nil {
		log.Fatalf("Start publicActor failed: %v", err)
	}

	// 启动GameServer
	if err := gs.Start(ctx); err != nil {
		log.Fatalf("Start GameServer failed: %v", err)
	}

	if err := dActor.Start(ctx); err != nil {
		log.Fatalf("Start DungeonActor failed: %v", err)
	}

	gevent.Publish(context.Background(), event.NewEvent(gevent.OnSrvStart))

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Infof("Shutting down GameServer...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := publicActor.Stop(shutdownCtx); err != nil {
		log.Errorf("Stop publicActor failed: %v", err)
	}
	if err := playerRoleActor.Stop(shutdownCtx); err != nil {
		log.Errorf("Stop playerRoleActor failed: %v", err)
	}
	if err := dActor.Stop(shutdownCtx); err != nil {
		log.Errorf("Stop DungeonActor failed: %v", err)
	}
	manager.GetPlayerRoleManager().FlushAndSave(shutdownCtx)
	if err := gs.Stop(shutdownCtx); err != nil {
		log.Fatalf("Stop GameServer failed: %v", err)
	}
	log.Infof("GameServer shutdown complete")
}
