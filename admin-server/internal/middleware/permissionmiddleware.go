package middleware

import (
	"context"
	"net/http"
	"strings"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"
	"postapocgame/admin-server/pkg/response"
)

// PermissionMiddleware 权限鉴权中间件
type PermissionMiddleware struct {
	svcCtx *svc.ServiceContext
}

func NewPermissionMiddleware(svcCtx *svc.ServiceContext) *PermissionMiddleware {
	return &PermissionMiddleware{svcCtx: svcCtx}
}

func (m *PermissionMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取当前用户信息
		user, ok := jwthelper.FromContext(r.Context())
		if !ok {
			response.ErrorCtx(r.Context(), w, errs.New(errs.CodeUnauthorized, "未登录或登录已过期"))
			return
		}

		// 超级管理员（user_id=1）拥有所有权限，直接通过
		if user.UserID == 1 {
			next(w, r)
			return
		}

		// 获取请求的方法和路径
		method := r.Method
		path := r.URL.Path

		// 移除前缀 /api/v1
		if strings.HasPrefix(path, "/api/v1") {
			path = path[len("/api/v1"):]
		}

		// 查找对应的接口
		apiRepo := repository.NewApiRepository(m.svcCtx.Repository)
		api, err := apiRepo.FindByMethodAndPath(r.Context(), method, path)
		if err != nil {
			// 如果接口不存在，可能是新接口还未配置，或者路径不匹配
			// 尝试匹配带参数的路径（如 /users/:id -> /users/123）
			matchedApi, matchErr := m.findApiByPattern(r.Context(), method, path)
			if matchErr != nil {
				// 接口未配置，允许通过（或者返回错误，根据业务需求）
				// 这里选择允许通过，避免影响未配置的接口
				next(w, r)
				return
			}
			api = matchedApi
		}

		// 如果接口未启用，直接通过（不进行权限检查）
		if api.Status != 1 {
			next(w, r)
			return
		}

		// 获取用户的所有权限
		userRoleRepo := repository.NewUserRoleRepository(m.svcCtx.Repository)
		roleIds, err := userRoleRepo.ListRoleIDsByUserID(r.Context(), user.UserID)
		if err != nil {
			response.ErrorCtx(r.Context(), w, errs.Wrap(errs.CodeInternalError, "获取用户角色失败", err))
			return
		}

		if len(roleIds) == 0 {
			response.ErrorCtx(r.Context(), w, errs.New(errs.CodeForbidden, "无权限访问"))
			return
		}

		// 获取所有角色拥有的权限
		rolePermissionRepo := repository.NewRolePermissionRepository(m.svcCtx.Repository)
		permissionIds := make(map[uint64]bool)
		for _, roleId := range roleIds {
			permIds, err := rolePermissionRepo.ListPermissionIDsByRoleID(r.Context(), roleId)
			if err != nil {
				continue
			}
			for _, permId := range permIds {
				permissionIds[permId] = true
			}
		}

		// 检查是否有超级权限（*）
		if permissionIds[1] {
			next(w, r)
			return
		}

		// 查找接口关联的权限
		permissionApiRepo := repository.NewPermissionApiRepository(m.svcCtx.Repository)

		// 获取该接口关联的所有权限ID
		apiPermissionIds, err := permissionApiRepo.ListPermissionIDsByApiID(r.Context(), api.Id)
		if err != nil {
			// 如果查询失败，允许通过（避免影响系统）
			next(w, r)
			return
		}

		// 检查用户是否有该接口的权限
		hasPermission := false
		for _, apiPermId := range apiPermissionIds {
			if permissionIds[apiPermId] {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			response.ErrorCtx(r.Context(), w, errs.New(errs.CodeForbidden, "无权限访问该接口"))
			return
		}

		next(w, r)
	}
}

// findApiByPattern 尝试通过模式匹配查找接口（处理路径参数）
func (m *PermissionMiddleware) findApiByPattern(ctx context.Context, method, path string) (*model.AdminApi, error) {
	// 获取所有接口
	apiRepo := repository.NewApiRepository(m.svcCtx.Repository)
	apis, _, err := apiRepo.FindPage(ctx, 1, 1000, "")
	if err != nil {
		return nil, err
	}

	// 遍历所有接口，尝试匹配
	for i := range apis {
		api := &apis[i]
		if api.Method != method {
			continue
		}
		// 移除 /api/v1 前缀进行匹配
		apiPath := api.Path
		if strings.HasPrefix(apiPath, "/api/v1") {
			apiPath = apiPath[len("/api/v1"):]
		}
		if m.matchPath(apiPath, path) {
			return api, nil
		}
	}

	return nil, errs.New(errs.CodeNotFound, "接口不存在")
}

// matchPath 匹配路径（处理 :id 等参数）
func (m *PermissionMiddleware) matchPath(pattern, path string) bool {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, patternPart := range patternParts {
		pathPart := pathParts[i]
		// 如果是参数（以:开头），跳过
		if strings.HasPrefix(patternPart, ":") {
			continue
		}
		// 精确匹配
		if patternPart != pathPart {
			return false
		}
	}

	return true
}
