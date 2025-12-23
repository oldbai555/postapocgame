package main

import (
	"context"
	"os"
	"os/signal"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/dungeonactor"
	engine2 "postapocgame/server/service/gameserver/internel/engine"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/playeractor"
	"postapocgame/server/service/gameserver/internel/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/playeractor/register"
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

	globalRuntime := deps.NewRuntime(
		deps.NewPlayerGateway(),
		deps.NewRoleRepository(),
		deps.NewNetworkGateway(),
		deps.NewDungeonServerGateway(),
	)

	serverConfig, err := engine2.LoadServerConfig("")
	if err != nil {
		log.Fatalf("err:%v", err)
		return
	}
	if serverConfig == nil {
		log.Fatalf("server config is nil")
		return
	}

	platformID := serverConfig.PlatformID
	srvID := serverConfig.SrvId
	gshare.SetPlatformId(platformID)
	gshare.SetSrvId(srvID)

	serverInfo, err := database.EnsureServerInfo(platformID, srvID)
	if err != nil {
		log.Fatalf("ensure server info failed: %v", err)
		return
	}
	if serverInfo == nil {
		log.Fatalf("server info is nil")
		return
	}
	gshare.SetOpenSrvTime(serverInfo.ServerOpenTimeAt)

	// 创建GameServer
	gs := engine2.NewGameServer(serverConfig)

	// 玩家消息处理
	playerRoleActor := playeractor.NewPlayerRoleActor(actor.ModePerKey)
	if err := playerRoleActor.Init(); err != nil {
		log.Fatalf("err:%v", err)
	}

	// 副本 / 战斗 DungeonActor（单 Actor，常驻运行）
	dActor := dungeonactor.NewDungeonActor(actor.ModeSingle)

	// 完成所有注册（依赖已装配的 facade）
	register.All(globalRuntime)
	dungeonactor.RegisterHandlers()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := playerRoleActor.Start(ctx); err != nil {
		log.Fatalf("Start playerRoleActor failed: %v", err)
	}

	// 启动GameServer
	if err := gs.Start(ctx); err != nil {
		log.Fatalf("Start GameServer failed: %v", err)
	}

	if err := dActor.Start(ctx); err != nil {
		log.Fatalf("Start DungeonActor failed: %v", err)
	}

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Infof("Shutting down GameServer...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := playerRoleActor.Stop(shutdownCtx); err != nil {
		log.Errorf("Stop playerRoleActor failed: %v", err)
	}
	if err := dActor.Stop(shutdownCtx); err != nil {
		log.Errorf("Stop DungeonActor failed: %v", err)
	}
	// 获取 PlayerRoleManager，并指定批次大小（每批 100 个角色）
	if err := deps.GetPlayerRoleManager().FlushAndSave(shutdownCtx, 100); err != nil {
		log.Errorf("FlushAndSave failed: %v", err)
	}
	if err := gs.Stop(shutdownCtx); err != nil {
		log.Fatalf("Stop GameServer failed: %v", err)
	}
	log.Infof("GameServer shutdown complete")
}
