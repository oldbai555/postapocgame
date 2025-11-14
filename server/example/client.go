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
	log.Infof("   集成测试客户端")
	log.Infof("===============================")

	clientMgr := NewClientManager(ctx)
	defer clientMgr.Stop()

	scenario := NewIntegrationScenario(ctx, clientMgr)
	if err := scenario.Run(); err != nil {
		log.Errorf("❌ 集成测试失败: %v", err)
	} else {
		log.Infof("✅ 集成测试成功")
	}

	log.Infof("\n按 Enter 退出客户端...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
