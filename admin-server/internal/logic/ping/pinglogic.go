// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ping

import (
	"context"
	"errors"
	"time"

	"postapocgame/admin-server/internal/consts"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	// 服务启动时间
	startTime = time.Now().Unix()
	// 服务版本
	version = "1.0.0"
)

type PingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PingLogic) Ping() (resp *types.PingResp, err error) {
	// 检查数据库连接状态
	databaseStatus := consts.StatusOK
	if err := l.checkDatabase(); err != nil {
		databaseStatus = consts.StatusError
		l.Errorf("数据库连接检查失败: %v", err)
	}

	// 检查Redis连接状态
	redisStatus := consts.StatusOK
	if err := l.checkRedis(); err != nil {
		redisStatus = consts.StatusError
		l.Errorf("Redis连接检查失败: %v", err)
	}

	// 计算运行时长
	uptime := time.Now().Unix() - startTime

	// 确定服务状态
	status := consts.StatusOK
	if databaseStatus == consts.StatusError || redisStatus == consts.StatusError {
		status = consts.StatusError
	}

	return &types.PingResp{
		Status:    status,
		Message:   consts.PingMessagePong,
		Database:  databaseStatus,
		Redis:     redisStatus,
		Version:   version,
		StartTime: startTime,
		Uptime:    uptime,
	}, nil
}

// checkDatabase 检查数据库连接
func (l *PingLogic) checkDatabase() error {
	// 执行一个简单的查询来检查数据库连接
	var result int
	err := l.svcCtx.Repository.DB.QueryRowCtx(l.ctx, &result, "SELECT 1")
	return err
}

// checkRedis 检查Redis连接
func (l *PingLogic) checkRedis() error {
	// 执行一个简单的命令来检查Redis连接
	// go-zero Redis Ping() 返回 bool，表示连接是否成功
	if !l.svcCtx.Repository.Redis.Ping() {
		return errors.New(consts.RedisPingFailedMessage)
	}
	return nil
}
