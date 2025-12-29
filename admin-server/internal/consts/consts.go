package consts

// 通用状态字符串
const (
	StatusOK    = "ok"
	StatusError = "error"
)

// Ping 接口相关常量
const (
	PingMessagePong = "pong"
)

// Redis 相关常量
const (
	RedisPingFailedMessage = "redis ping failed"

	// JWT 黑名单前缀
	RedisJWTBlacklistPrefix = "jwt:blacklist:"

	// 限流相关 Redis 前缀
	RedisRateLimitGlobalPrefix = "rate_limit:global"
	RedisRateLimitIPPrefix     = "rate_limit:ip:"
	RedisRateLimitUserPrefix   = "rate_limit:user:"
	RedisRateLimitAPIPrefix    = "rate_limit:api:"
)

// 限流提示信息
const (
	RateLimitMessageGlobal = "请求过于频繁，请稍后再试（全局限流）"
	RateLimitMessageIP     = "请求过于频繁，请稍后再试（IP限流）"
	RateLimitMessageUser   = "请求过于频繁，请稍后再试（用户限流）"
	RateLimitMessageAPI    = "请求过于频繁，请稍后再试（接口限流）"
)

// 常用路径常量
const (
	PathPing = "/api/v1/ping"

	// 认证相关路径
	PathLogin   = "/api/v1/login"
	PathLogout  = "/api/v1/logout"
	PathRefresh = "/api/v1/refresh"

	// WebSocket 路径
	PathChatWS = "/api/v1/chats/ws"
)
