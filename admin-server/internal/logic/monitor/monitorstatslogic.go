// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package monitor

import (
	"context"

	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MonitorStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMonitorStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MonitorStatsLogic {
	return &MonitorStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MonitorStatsLogic) MonitorStats() (resp *types.MonitorStatsResp, err error) {
	// 统计用户数
	userCount, err := l.countUsers()
	if err != nil {
		l.Errorf("统计用户数失败: %v", err)
		userCount = 0
	}

	// 统计角色数
	roleCount, err := l.countRoles()
	if err != nil {
		l.Errorf("统计角色数失败: %v", err)
		roleCount = 0
	}

	// 统计权限数
	permissionCount, err := l.countPermissions()
	if err != nil {
		l.Errorf("统计权限数失败: %v", err)
		permissionCount = 0
	}

	// 统计菜单数
	menuCount, err := l.countMenus()
	if err != nil {
		l.Errorf("统计菜单数失败: %v", err)
		menuCount = 0
	}

	// 统计在线用户数（从 WebSocket Hub 获取）
	// 注意：ChatOnlineUser 表已移除，在线用户数从 WebSocket Hub 获取
	onlineUserCount := int64(0)
	if l.svcCtx.ChatHub != nil {
		onlineUserIDs := l.svcCtx.ChatHub.GetOnlineUsers()
		onlineUserCount = int64(len(onlineUserIDs))
	}

	// 统计操作日志数
	operationLogCount, err := l.countOperationLogs()
	if err != nil {
		l.Errorf("统计操作日志数失败: %v", err)
		operationLogCount = 0
	}

	// 统计登录日志数
	loginLogCount, err := l.countLoginLogs()
	if err != nil {
		l.Errorf("统计登录日志数失败: %v", err)
		loginLogCount = 0
	}

	return &types.MonitorStatsResp{
		UserCount:         userCount,
		RoleCount:         roleCount,
		PermissionCount:   permissionCount,
		MenuCount:         menuCount,
		OnlineUserCount:   onlineUserCount,
		OperationLogCount: operationLogCount,
		LoginLogCount:     loginLogCount,
	}, nil
}

// countUsers 统计用户数
func (l *MonitorStatsLogic) countUsers() (int64, error) {
	var count int64
	err := l.svcCtx.Repository.DB.QueryRowCtx(l.ctx, &count, "SELECT COUNT(*) FROM `admin_user` WHERE deleted_at = 0")
	return count, err
}

// countRoles 统计角色数
func (l *MonitorStatsLogic) countRoles() (int64, error) {
	var count int64
	err := l.svcCtx.Repository.DB.QueryRowCtx(l.ctx, &count, "SELECT COUNT(*) FROM `admin_role` WHERE deleted_at = 0")
	return count, err
}

// countPermissions 统计权限数
func (l *MonitorStatsLogic) countPermissions() (int64, error) {
	var count int64
	err := l.svcCtx.Repository.DB.QueryRowCtx(l.ctx, &count, "SELECT COUNT(*) FROM `admin_permission` WHERE deleted_at = 0")
	return count, err
}

// countMenus 统计菜单数
func (l *MonitorStatsLogic) countMenus() (int64, error) {
	var count int64
	err := l.svcCtx.Repository.DB.QueryRowCtx(l.ctx, &count, "SELECT COUNT(*) FROM `admin_menu` WHERE deleted_at = 0")
	return count, err
}

// countOperationLogs 统计操作日志数
func (l *MonitorStatsLogic) countOperationLogs() (int64, error) {
	var count int64
	err := l.svcCtx.Repository.DB.QueryRowCtx(l.ctx, &count, "SELECT COUNT(*) FROM `admin_operation_log` WHERE deleted_at = 0")
	return count, err
}

// countLoginLogs 统计登录日志数
func (l *MonitorStatsLogic) countLoginLogs() (int64, error) {
	var count int64
	err := l.svcCtx.Repository.DB.QueryRowCtx(l.ctx, &count, "SELECT COUNT(*) FROM `admin_login_log` WHERE deleted_at = 0")
	return count, err
}
