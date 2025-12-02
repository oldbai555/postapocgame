package context

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/core/gshare"
	"postapocgame/server/service/gameserver/internel/core/iface"
)

// GetPlayerRoleFromContext 从 Context 获取 PlayerRole
func GetPlayerRoleFromContext(ctx context.Context) (iface.IPlayerRole, error) {
	value := ctx.Value(gshare.ContextKeyRole)
	if value == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "not found %s value", gshare.ContextKeyRole)
	}
	playerRole, ok := value.(iface.IPlayerRole)
	if !ok {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "invalid role type")
	}
	return playerRole, nil
}

// GetSessionIDFromContext 从 Context 获取 SessionID
func GetSessionIDFromContext(ctx context.Context) (string, error) {
	playerRole, err := GetPlayerRoleFromContext(ctx)
	if err != nil {
		return "", err
	}
	return playerRole.GetSessionId(), nil
}

// GetRoleIDFromContext 从 Context 获取 RoleID
func GetRoleIDFromContext(ctx context.Context) (uint64, error) {
	playerRole, err := GetPlayerRoleFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return playerRole.GetPlayerRoleId(), nil
}

// MustGetPlayerRoleFromContext 从 Context 获取 PlayerRole（如果失败则记录错误并返回 nil）
func MustGetPlayerRoleFromContext(ctx context.Context) iface.IPlayerRole {
	playerRole, err := GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role from context error: %v", err)
		return nil
	}
	return playerRole
}

// MustGetSessionIDFromContext 从 Context 获取 SessionID（如果失败则记录错误并返回空字符串）
func MustGetSessionIDFromContext(ctx context.Context) string {
	sessionID, err := GetSessionIDFromContext(ctx)
	if err != nil {
		log.Errorf("get session id from context error: %v", err)
		return ""
	}
	return sessionID
}

// MustGetRoleIDFromContext 从 Context 获取 RoleID（如果失败则记录错误并返回 0）
func MustGetRoleIDFromContext(ctx context.Context) uint64 {
	roleID, err := GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("get role id from context error: %v", err)
		return 0
	}
	return roleID
}
