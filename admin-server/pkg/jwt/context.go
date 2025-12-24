package jwt

import "context"

type ctxKey string

const userCtxKey ctxKey = "authUser"

// AuthUser 放入上下文的认证用户信息。
type AuthUser struct {
	UserID   uint64
	Username string
}

// WithAuthUser 将登录用户信息写入 context。
func WithAuthUser(ctx context.Context, user AuthUser) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

// FromContext 从 context 中读取登录用户信息。
func FromContext(ctx context.Context) (AuthUser, bool) {
	if ctx == nil {
		return AuthUser{}, false
	}
	val := ctx.Value(userCtxKey)
	if val == nil {
		return AuthUser{}, false
	}
	if u, ok := val.(AuthUser); ok {
		return u, true
	}
	return AuthUser{}, false
}
