package main

import (
	"fmt"
	"os"
	"postapocgame/server/internal/database"
)

func main() {
	dbPath := "postapocgame.db"
	if err := database.Init(dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "数据库初始化失败: %v\n", err)
		os.Exit(1)
	}
	if err := database.AutoMigrate(); err != nil {
		fmt.Fprintf(os.Stderr, "数据表自动迁移失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("数据库初始化和表迁移成功！")
	// 这里后续可添加集成服务或测试等入口...
}
