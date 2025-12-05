package gshare

import (
	"context"
	"fmt"
	"postapocgame/server/service/gameserver/internel/iface"
	"strings"

	"postapocgame/server/pkg/log"
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
	return "[" + strings.Join(parts, " ") + "]"
}

func InfofCtx(ctx context.Context, format string, v ...interface{}) {
	log.InfofWithRequester(newCtxRequester(ctx), format, v...)
}

func WarnfCtx(ctx context.Context, format string, v ...interface{}) {
	log.WarnfWithRequester(newCtxRequester(ctx), format, v...)
}

func ErrorfCtx(ctx context.Context, format string, v ...interface{}) {
	log.ErrorfWithRequester(newCtxRequester(ctx), format, v...)
}

func newCtxRequester(ctx context.Context) log.IRequester {
	base := log.GetSkipCall()
	skip := log.DefaultSkipCall
	if base != nil {
		skip = base.GetLogCallStackSkip()
	}
	// +1 to skip the gshare helper wrapper itself
	return log.NewRequester(buildLogPrefix(ctx), skip+1)
}
