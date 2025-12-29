package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	jwthelper "postapocgame/admin-server/pkg/jwt"
	"postapocgame/admin-server/pkg/monitor"

	"github.com/zeromicro/go-zero/core/logx"
)

// PerformanceMiddleware 接口性能监控中间件
type PerformanceMiddleware struct {
	svcCtx        *svc.ServiceContext
	monitor       *monitor.PerformanceMonitor
	slowThreshold int64 // 慢接口阈值（毫秒）
}

// NewPerformanceMiddleware 创建接口性能监控中间件
func NewPerformanceMiddleware(svcCtx *svc.ServiceContext) *PerformanceMiddleware {
	slowThreshold := int64(2000) // 默认 2 秒
	if svcCtx.Config.Database.SlowQueryThreshold > 0 {
		// 使用数据库慢查询阈值的 2 倍作为接口慢查询阈值
		slowThreshold = int64(svcCtx.Config.Database.SlowQueryThreshold) * 2
	}

	return &PerformanceMiddleware{
		svcCtx:        svcCtx,
		monitor:       monitor.NewPerformanceMonitor(slowThreshold),
		slowThreshold: slowThreshold,
	}
}

// Handle 中间件处理函数
func (m *PerformanceMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 记录开始时间
		startTime := time.Now()

		// 包装 ResponseWriter 以捕获状态码
		responseWriter := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           nil, // 性能监控不需要响应体
		}

		// 执行下一个处理器
		next(responseWriter, r)

		// 计算耗时
		duration := time.Since(startTime)

		// 记录接口性能
		m.monitor.RecordAPICall(
			r.Context(),
			r.Method,
			r.URL.Path,
			duration,
			responseWriter.statusCode,
			nil, // 错误信息从响应状态码判断
		)

		// 仅在慢接口或错误时写入性能日志表，避免数据量过大
		durationMs := duration.Milliseconds()
		isSlow := int64(0)
		if durationMs >= m.slowThreshold {
			isSlow = 1
		}

		if isSlow == 0 && responseWriter.statusCode < http.StatusBadRequest {
			// 非慢接口且未出错，不写入性能日志表
			return
		}

		// 获取用户信息
		user, ok := jwthelper.FromContext(r.Context())
		userId := uint64(0)
		username := ""
		if ok {
			userId = user.UserID
			username = user.Username
		}

		// 错误信息（状态码 >= 400）
		errorMsg := ""
		if responseWriter.statusCode >= http.StatusBadRequest {
			errorMsg = http.StatusText(responseWriter.statusCode)
		}

		// 异步写入性能日志表
		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					logx.Errorf("写入性能监控日志时发生 panic: %v", rec)
				}
			}()

			performanceLogRepo := repository.NewPerformanceLogRepository(m.svcCtx.Repository)
			now := time.Now().Unix()
			log := &model.AdminPerformanceLog{
				UserId:        userId,
				Username:      username,
				Method:        r.Method,
				Path:          r.URL.Path,
				StatusCode:    int64(responseWriter.statusCode),
				Duration:      durationMs,
				IsSlow:        isSlow,
				SlowThreshold: m.slowThreshold,
				IpAddress:     getClientIP(r),
				UserAgent:     r.UserAgent(),
				ErrorMsg:      errorMsg,
				CreatedAt:     now,
				UpdatedAt:     now,
				DeletedAt:     0,
			}

			if err := performanceLogRepo.Create(context.Background(), log); err != nil {
				logx.Errorf("写入性能监控日志失败: method=%s, path=%s, duration=%dms, status=%d, error=%v",
					r.Method, r.URL.Path, durationMs, responseWriter.statusCode, err)
			}
		}()
	}
}

// getClientIP 获取客户端 IP 地址
func getClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	// 优先从 X-Forwarded-For 获取（代理场景）
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 其次从 X-Real-IP 获取
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// 最后从 RemoteAddr 获取
	ip = r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}
