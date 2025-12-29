package middleware

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"postapocgame/admin-server/internal/consts"
	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

// OperationLogMiddleware 操作日志中间件，自动记录所有增删改操作的日志
type OperationLogMiddleware struct {
	svcCtx *svc.ServiceContext
	logCh  chan *model.AdminOperationLog // 异步日志通道
}

func NewOperationLogMiddleware(svcCtx *svc.ServiceContext) *OperationLogMiddleware {
	m := &OperationLogMiddleware{
		svcCtx: svcCtx,
		logCh:  make(chan *model.AdminOperationLog, 1000), // 缓冲1000条日志
	}

	// 启动异步日志写入 goroutine
	go m.logWriter()

	return m
}

// Handle 中间件处理函数
func (m *OperationLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 只记录增删改操作（POST、PUT、DELETE）
		method := r.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete {
			next(w, r)
			return
		}

		// 排除一些不需要记录的接口
		path := r.URL.Path
		if m.shouldSkip(path) {
			next(w, r)
			return
		}

		// 记录开始时间
		startTime := time.Now()

		// 读取请求体（用于记录请求参数）
		var requestBody []byte
		if r.Body != nil {
			requestBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody)) // 恢复 body，供后续处理使用
		}

		// 包装 ResponseWriter 以捕获响应
		responseWriter := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           &bytes.Buffer{},
		}

		// 执行下一个处理器
		next(responseWriter, r)

		// 计算耗时
		duration := int(time.Since(startTime).Milliseconds())

		// 获取用户信息
		user, ok := jwthelper.FromContext(r.Context())
		userId := uint64(0)
		username := ""
		if ok {
			userId = user.UserID
			username = user.Username
		}

		// 解析操作类型和操作对象
		operationType, operationObject := m.parseOperation(method, path)

		// 构建操作日志
		requestParams := sql.NullString{}
		if len(requestBody) > 0 {
			// 限制请求参数长度，避免过长
			paramsStr := string(requestBody)
			if len(paramsStr) > 10000 {
				paramsStr = paramsStr[:10000] + "..."
			}
			requestParams = sql.NullString{String: paramsStr, Valid: true}
		}

		operationLog := &model.AdminOperationLog{
			UserId:          userId,
			Username:        username,
			OperationType:   operationType,
			OperationObject: operationObject,
			Method:          method,
			Path:            path,
			RequestParams:   requestParams,
			ResponseCode:    int64(responseWriter.statusCode),
			ResponseMsg:     m.extractResponseMsg(responseWriter.body.String()),
			IpAddress:       m.getClientIP(r),
			UserAgent:       r.UserAgent(),
			Duration:        int64(duration),
			DeletedAt:       0, // 软删除字段，0 表示未删除
		}

		// 异步写入日志（非阻塞）
		select {
		case m.logCh <- operationLog:
			logx.Infof("操作日志已加入队列: method=%s, path=%s, userId=%d", method, path, userId)
		default:
			// 通道满了，记录警告但不阻塞请求
			logx.Errorf("操作日志通道已满，丢弃日志: %+v", operationLog)
		}
	}
}

// shouldSkip 判断是否应该跳过记录
func (m *OperationLogMiddleware) shouldSkip(path string) bool {
	skipPaths := []string{
		consts.PathPing,
		consts.PathLogin,
		consts.PathRefresh,
		consts.PathLogout,
		consts.PathChatWS, // WebSocket 连接
	}
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// parseOperation 解析操作类型和操作对象
func (m *OperationLogMiddleware) parseOperation(method, path string) (operationType, operationObject string) {
	// 根据 HTTP 方法确定操作类型
	switch method {
	case http.MethodPost:
		operationType = "create"
	case http.MethodPut:
		operationType = "update"
	case http.MethodDelete:
		operationType = "delete"
	default:
		operationType = "unknown"
	}

	// 从路径中提取操作对象（模块名）
	// 例如：/api/v1/users -> user, /api/v1/roles -> role
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 {
		// 移除 /api/v1 前缀
		module := parts[2]
		// 移除复数形式（如 users -> user）
		if strings.HasSuffix(module, "s") && len(module) > 1 {
			module = module[:len(module)-1]
		}
		operationObject = module
	}

	return operationType, operationObject
}

// extractResponseMsg 从响应体中提取消息
func (m *OperationLogMiddleware) extractResponseMsg(responseBody string) string {
	if responseBody == "" {
		return ""
	}

	// 尝试解析 JSON 响应
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(responseBody), &resp); err == nil {
		if msg, ok := resp["msg"].(string); ok {
			return msg
		}
	}

	// 如果解析失败，返回前255个字符
	if len(responseBody) > 255 {
		return responseBody[:255]
	}
	return responseBody
}

// getClientIP 获取客户端 IP 地址
func (m *OperationLogMiddleware) getClientIP(r *http.Request) string {
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

// logWriter 异步日志写入器
func (m *OperationLogMiddleware) logWriter() {
	operationLogRepo := repository.NewOperationLogRepository(m.svcCtx.Repository)
	batch := make([]*model.AdminOperationLog, 0, 100) // 批量写入，每批100条
	ticker := time.NewTicker(5 * time.Second)         // 每5秒写入一次
	defer ticker.Stop()

	for {
		select {
		case log := <-m.logCh:
			batch = append(batch, log)
			// 如果批次达到100条，立即写入
			if len(batch) >= 100 {
				m.writeBatch(operationLogRepo, batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			// 定时写入批次中的日志
			if len(batch) > 0 {
				m.writeBatch(operationLogRepo, batch)
				batch = batch[:0]
			}
		}
	}
}

// writeBatch 批量写入日志
func (m *OperationLogMiddleware) writeBatch(repo repository.OperationLogRepository, logs []*model.AdminOperationLog) {
	ctx := context.Background()
	if len(logs) == 0 {
		return
	}

	// 使用批量创建方法
	if err := repo.BatchCreate(ctx, logs); err != nil {
		logx.Errorf("批量写入操作日志失败: count=%d, error: %v", len(logs), err)
		// 如果批量写入失败，尝试逐条写入
		for _, log := range logs {
			if err := repo.Create(ctx, log); err != nil {
				logx.Errorf("写入操作日志失败: %+v, error: %v", log, err)
			}
		}
	} else {
		logx.Infof("成功批量写入操作日志: count=%d", len(logs))
	}
}

// responseWriterWrapper 包装 ResponseWriter 以捕获响应
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	// body 可能为 nil（例如仅需记录状态码的中间件场景），需要判空避免 panic
	if w.body != nil {
		w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}
