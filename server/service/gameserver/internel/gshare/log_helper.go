package gshare

import (
	"context"
	"fmt"
	"strings"

	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
)

func buildLogPrefix(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	var parts []string
	if sessionID, ok := ctx.Value(ContextKeySession).(string); ok && sessionID != "" {
		parts = append(parts, fmt.Sprintf("session=%s", sessionID))
	}

	if roleVal := ctx.Value(ContextKeyRole); roleVal != nil {
		switch v := roleVal.(type) {
		case iface.IPlayerRole:
			parts = append(parts, fmt.Sprintf("role=%d", v.GetPlayerRoleId()))
		case interface{ GetPlayerRoleId() uint64 }:
			parts = append(parts, fmt.Sprintf("role=%d", v.GetPlayerRoleId()))
		case uint64:
			parts = append(parts, fmt.Sprintf("role=%d", v))
		case int64:
			parts = append(parts, fmt.Sprintf("role=%d", v))
		case int:
			parts = append(parts, fmt.Sprintf("role=%d", v))
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return "[" + strings.Join(parts, " ") + "] "
}

func InfofCtx(ctx context.Context, format string, v ...interface{}) {
	log.Infof(buildLogPrefix(ctx)+format, v...)
}

func WarnfCtx(ctx context.Context, format string, v ...interface{}) {
	log.Warnf(buildLogPrefix(ctx)+format, v...)
}

func ErrorfCtx(ctx context.Context, format string, v ...interface{}) {
	log.Errorf(buildLogPrefix(ctx)+format, v...)
}
