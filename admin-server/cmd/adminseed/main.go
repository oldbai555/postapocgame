package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"postapocgame/admin-server/internal/config"
	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
)

// 一个简单的种子工具：根据配置连接数据库，创建默认管理员账号和基础角色/权限。
// 使用方式（在 admin-server 目录）：go run ./cmd/adminseed -f etc/admin-api.yaml -username admin -password 123456

var (
	configFile = flag.String("f", "etc/admin-api.yaml", "the config file")
	username   = flag.String("username", "oldbai", "admin username")
	password   = flag.String("password", "oldbai", "admin password (will be bcrypt hashed)")
)

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	repo, err := repository.BuildSources(c)
	if err != nil {
		log.Fatalf("init repository failed: %v", err)
	}

	ctx := context.Background()
	userModel := repo.AdminUserModel

	// 生成密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(*password), c.Bcrypt.Cost)
	if err != nil {
		log.Fatalf("generate password hash failed: %v", err)
	}

	// 创建管理员用户（若不存在）
	user, err := userModel.FindOneByUsername(ctx, *username)
	if err == nil && user != nil {
		fmt.Printf("Admin user already exists: username=%s\n", *username)
		return
	}
	if err != nil && err != sqlc.ErrNotFound {
		log.Fatalf("query admin user failed: %v", err)
	}

	_, err = userModel.Insert(ctx, &model.AdminUser{
		Username:     *username,
		PasswordHash: string(hash),
		Status:       1,
	})
	if err != nil {
		log.Fatalf("create admin user failed: %v", err)
	}
	fmt.Printf("Admin user created: username=%s\n", *username)
}
