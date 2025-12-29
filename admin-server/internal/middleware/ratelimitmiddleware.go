package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"postapocgame/admin-server/internal/consts"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"
	"postapocgame/admin-server/pkg/response"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
)

// RateLimitMiddleware 限流中间件，支持按IP、按用户、按接口限流
type RateLimitMiddleware struct {
	svcCtx *svc.ServiceContext
	// 按IP限流器映射
	ipLimiters map[string]*limit.PeriodLimit
	// 按用户限流器映射
	userLimiters map[uint64]*limit.PeriodLimit
	// 按接口限流器映射
	apiLimiters map[string]*limit.PeriodLimit
	// 全局限流器（所有请求共享）
	globalLimiter *limit.PeriodLimit
	// 保护并发访问的锁
	mu sync.RWMutex
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 是否启用限流
	Enabled bool `json:"enabled" yaml:"enabled"`
	// 按IP限流配置
	IPLimit struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
		Quota   int  `json:"quota" yaml:"quota"`   // 时间窗口内的请求数
		Period  int  `json:"period" yaml:"period"` // 时间窗口（秒）
	} `json:"ipLimit" yaml:"ipLimit"`
	// 按用户限流配置
	UserLimit struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
		Quota   int  `json:"quota" yaml:"quota"`   // 时间窗口内的请求数
		Period  int  `json:"period" yaml:"period"` // 时间窗口（秒）
	} `json:"userLimit" yaml:"userLimit"`
	// 按接口限流配置
	APILimit struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
		Quota   int  `json:"quota" yaml:"quota"`   // 时间窗口内的请求数
		Period  int  `json:"period" yaml:"period"` // 时间窗口（秒）
	} `json:"apiLimit" yaml:"apiLimit"`
	// 全局限流配置
	GlobalLimit struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
		Quota   int  `json:"quota" yaml:"quota"`   // 时间窗口内的请求数
		Period  int  `json:"period" yaml:"period"` // 时间窗口（秒）
	} `json:"globalLimit" yaml:"globalLimit"`
}

func NewRateLimitMiddleware(svcCtx *svc.ServiceContext) *RateLimitMiddleware {
	m := &RateLimitMiddleware{
		svcCtx:       svcCtx,
		ipLimiters:   make(map[string]*limit.PeriodLimit),
		userLimiters: make(map[uint64]*limit.PeriodLimit),
		apiLimiters:  make(map[string]*limit.PeriodLimit),
	}

	// 从配置中读取限流规则
	config := m.getRateLimitConfig()

	// 初始化全局限流器
	if config.GlobalLimit.Enabled {
		m.globalLimiter = limit.NewPeriodLimit(
			config.GlobalLimit.Period,
			config.GlobalLimit.Quota,
			svcCtx.Repository.Redis,
			consts.RedisRateLimitGlobalPrefix,
		)
	}

	return m
}

// getRateLimitConfig 获取限流配置（从配置文件或默认值）
func (m *RateLimitMiddleware) getRateLimitConfig() RateLimitConfig {
	// 从配置文件读取限流配置
	cfg := m.svcCtx.Config.RateLimit

	// 如果配置文件中没有配置，使用默认值
	config := RateLimitConfig{
		Enabled: cfg.Enabled,
		IPLimit: struct {
			Enabled bool `json:"enabled" yaml:"enabled"`
			Quota   int  `json:"quota" yaml:"quota"`
			Period  int  `json:"period" yaml:"period"`
		}{
			Enabled: cfg.IPLimit.Enabled,
			Quota:   cfg.IPLimit.Quota,
			Period:  cfg.IPLimit.Period,
		},
		UserLimit: struct {
			Enabled bool `json:"enabled" yaml:"enabled"`
			Quota   int  `json:"quota" yaml:"quota"`
			Period  int  `json:"period" yaml:"period"`
		}{
			Enabled: cfg.UserLimit.Enabled,
			Quota:   cfg.UserLimit.Quota,
			Period:  cfg.UserLimit.Period,
		},
		APILimit: struct {
			Enabled bool `json:"enabled" yaml:"enabled"`
			Quota   int  `json:"quota" yaml:"quota"`
			Period  int  `json:"period" yaml:"period"`
		}{
			Enabled: cfg.APILimit.Enabled,
			Quota:   cfg.APILimit.Quota,
			Period:  cfg.APILimit.Period,
		},
		GlobalLimit: struct {
			Enabled bool `json:"enabled" yaml:"enabled"`
			Quota   int  `json:"quota" yaml:"quota"`
			Period  int  `json:"period" yaml:"period"`
		}{
			Enabled: cfg.GlobalLimit.Enabled,
			Quota:   cfg.GlobalLimit.Quota,
			Period:  cfg.GlobalLimit.Period,
		},
	}

	// 如果配置未启用，使用默认值
	if !config.Enabled {
		config.Enabled = true
	}
	if config.IPLimit.Quota == 0 {
		config.IPLimit.Enabled = true
		config.IPLimit.Quota = 100
		config.IPLimit.Period = 1
	}
	if config.UserLimit.Quota == 0 {
		config.UserLimit.Enabled = true
		config.UserLimit.Quota = 200
		config.UserLimit.Period = 1
	}
	if config.APILimit.Quota == 0 {
		config.APILimit.Enabled = true
		config.APILimit.Quota = 50
		config.APILimit.Period = 1
	}
	if config.GlobalLimit.Quota == 0 {
		config.GlobalLimit.Enabled = true
		config.GlobalLimit.Quota = 1000
		config.GlobalLimit.Period = 1
	}

	return config
}

// getIPLimiter 获取或创建IP限流器
func (m *RateLimitMiddleware) getIPLimiter(ip string, quota, period int) *limit.PeriodLimit {
	m.mu.RLock()
	limiter, exists := m.ipLimiters[ip]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		defer m.mu.Unlock()
		// 双重检查
		if limiter, exists = m.ipLimiters[ip]; !exists {
			limiter = limit.NewPeriodLimit(
				period,
				quota,
				m.svcCtx.Repository.Redis,
				consts.RedisRateLimitIPPrefix+ip,
			)
			m.ipLimiters[ip] = limiter
		}
	}

	return limiter
}

// getUserLimiter 获取或创建用户限流器
func (m *RateLimitMiddleware) getUserLimiter(userID uint64, quota, period int) *limit.PeriodLimit {
	m.mu.RLock()
	limiter, exists := m.userLimiters[userID]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		defer m.mu.Unlock()
		// 双重检查
		if limiter, exists = m.userLimiters[userID]; !exists {
			limiter = limit.NewPeriodLimit(
				period,
				quota,
				m.svcCtx.Repository.Redis,
				consts.RedisRateLimitUserPrefix+strconv.FormatUint(userID, 10),
			)
			m.userLimiters[userID] = limiter
		}
	}

	return limiter
}

// getAPILimiter 获取或创建接口限流器
func (m *RateLimitMiddleware) getAPILimiter(apiKey string, quota, period int) *limit.PeriodLimit {
	m.mu.RLock()
	limiter, exists := m.apiLimiters[apiKey]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		defer m.mu.Unlock()
		// 双重检查
		if limiter, exists = m.apiLimiters[apiKey]; !exists {
			limiter = limit.NewPeriodLimit(
				period,
				quota,
				m.svcCtx.Repository.Redis,
				consts.RedisRateLimitAPIPrefix+apiKey,
			)
			m.apiLimiters[apiKey] = limiter
		}
	}

	return limiter
}

// getClientIP 获取客户端IP地址
func (m *RateLimitMiddleware) getClientIP(r *http.Request) string {
	// 优先从 X-Forwarded-For 获取（代理场景）
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 其次从 X-Real-IP 获取
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// 最后从 RemoteAddr 获取
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// Handle 中间件处理函数
func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config := m.getRateLimitConfig()

		// 如果限流未启用，直接通过
		if !config.Enabled {
			next(w, r)
			return
		}

		// 排除一些不需要限流的接口
		path := r.URL.Path
		if m.shouldSkip(path) {
			next(w, r)
			return
		}

		// 1. 全局限流检查
		if config.GlobalLimit.Enabled && m.globalLimiter != nil {
			code, err := m.globalLimiter.Take(consts.RedisRateLimitGlobalPrefix)
			if err != nil {
				logx.Errorf("全局限流检查失败: %v", err)
			} else if code == limit.OverQuota {
				m.rateLimitResponse(w, r, consts.RateLimitMessageGlobal)
				return
			}
		}

		// 2. 按IP限流检查
		if config.IPLimit.Enabled {
			ip := m.getClientIP(r)
			limiter := m.getIPLimiter(ip, config.IPLimit.Quota, config.IPLimit.Period)
			code, err := limiter.Take(ip)
			if err != nil {
				logx.Errorf("IP限流检查失败: %v", err)
			} else if code == limit.OverQuota {
				logx.Infof("IP限流触发: IP=%s, Path=%s", ip, path)
				m.rateLimitResponse(w, r, consts.RateLimitMessageIP)
				return
			}
		}

		// 3. 按用户限流检查（需要用户已登录）
		if config.UserLimit.Enabled {
			user, ok := jwthelper.FromContext(r.Context())
			if ok {
				limiter := m.getUserLimiter(user.UserID, config.UserLimit.Quota, config.UserLimit.Period)
				code, err := limiter.Take(strconv.FormatUint(user.UserID, 10))
				if err != nil {
					logx.Errorf("用户限流检查失败: %v", err)
				} else if code == limit.OverQuota {
					logx.Infof("用户限流触发: UserID=%d, Username=%s, Path=%s", user.UserID, user.Username, path)
					m.rateLimitResponse(w, r, consts.RateLimitMessageUser)
					return
				}
			}
		}

		// 4. 按接口限流检查
		if config.APILimit.Enabled {
			apiKey := r.Method + ":" + path
			limiter := m.getAPILimiter(apiKey, config.APILimit.Quota, config.APILimit.Period)
			code, err := limiter.Take(apiKey)
			if err != nil {
				logx.Errorf("接口限流检查失败: %v", err)
			} else if code == limit.OverQuota {
				logx.Infof("接口限流触发: API=%s", apiKey)
				m.rateLimitResponse(w, r, consts.RateLimitMessageAPI)
				return
			}
		}

		// 所有限流检查通过，继续处理请求
		next(w, r)
	}
}

// shouldSkip 判断是否跳过限流检查
func (m *RateLimitMiddleware) shouldSkip(path string) bool {
	// 排除健康检查接口
	if path == consts.PathPing {
		return true
	}
	// 排除登录接口（登录接口通常需要更严格的限流，但这里先跳过）
	// 如果需要，可以在登录接口单独实现限流逻辑
	return false
}

// rateLimitResponse 返回限流响应
func (m *RateLimitMiddleware) rateLimitResponse(w http.ResponseWriter, r *http.Request, message string) {
	// 设置HTTP状态码为429
	w.WriteHeader(http.StatusTooManyRequests)
	// 返回限流错误响应
	response.ErrorCtx(r.Context(), w, errs.New(http.StatusTooManyRequests, message))
}
