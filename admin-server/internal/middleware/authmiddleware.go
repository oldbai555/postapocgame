package middleware

import (
	"net/http"
	"strings"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"
	"postapocgame/admin-server/pkg/response"
)

// AuthMiddleware 校验 Access Token + 黑名单，并将用户信息写入 context。
type AuthMiddleware struct {
	svcCtx *svc.ServiceContext
}

func NewAuthMiddleware(svcCtx *svc.ServiceContext) *AuthMiddleware {
	return &AuthMiddleware{svcCtx: svcCtx}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.ErrorCtx(r.Context(), w, errs.New(errs.CodeUnauthorized, "未提供认证信息"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			response.ErrorCtx(r.Context(), w, errs.New(errs.CodeUnauthorized, "无效的认证头"))
			return
		}
		token := parts[1]

		// 黑名单校验
		blackRepo := repository.NewTokenBlacklistRepository(m.svcCtx.Repository)
		blacklisted, err := blackRepo.IsBlacklisted(r.Context(), token)
		if err != nil {
			response.ErrorCtx(r.Context(), w, errs.Wrap(errs.CodeInternalError, "检查令牌黑名单失败", err))
			return
		}
		if blacklisted {
			response.ErrorCtx(r.Context(), w, errs.New(errs.CodeUnauthorized, "令牌已失效"))
			return
		}

		// 解析 Access Token
		claims, err := jwthelper.ParseToken(token, m.svcCtx.Config.JWT.AccessSecret)
		if err != nil || claims.IsRefresh {
			response.ErrorCtx(r.Context(), w, errs.New(errs.CodeUnauthorized, "访问令牌无效或已过期"))
			return
		}

		ctxWithUser := jwthelper.WithAuthUser(r.Context(), jwthelper.AuthUser{
			UserID:   claims.UserID,
			Username: claims.Username,
		})

		next(w, r.WithContext(ctxWithUser))
	}
}
