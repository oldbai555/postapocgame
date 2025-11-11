package main

import (
	"bufio"
	"context"
	"os"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
)

func main() {
	log.InitLogger(log.WithAppName("example_client_actor"), log.WithScreen(true), log.WithPath(tool.GetCurDir()+"log"))
	defer log.Flush()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Infof("===============================")
	log.Infof("   游戏客户端测试程序 (Actor模式)")
	log.Infof("===============================")

	// 创建客户端管理器
	clientMgr := NewClientManager(ctx)

	// 创建玩家客户端
	player1 := clientMgr.CreateClient("player1", GatewayAddr)

	// 启动客户端
	if err := player1.Start(ctx); err != nil {
		log.Errorf("❌ Player1 启动失败: %v\n", err)
		return
	}
	log.Infof("✅ Player1 已连接")

	// 查询角色
	log.Infof("\n[Player1] 查询角色列表...")
	if err := player1.QueryRoles(); err != nil {
		log.Errorf("❌ Player1 查询角色失败: %v\n", err)
	}

	log.Infof("\n✅ 测试完成，按 Enter 退出...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	// 停止客户端管理器
	clientMgr.Stop()
}
