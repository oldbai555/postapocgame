// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"
	"postapocgame/admin-server/pkg/useragent"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq, httpReq *http.Request) (resp *types.TokenPair, err error) {
	if req == nil || req.Username == "" || req.Password == "" {
		// 记录登录失败日志
		l.recordLoginLog(0, req.Username, httpReq, "用户名和密码不能为空", false)
		return nil, errs.New(errs.CodeBadRequest, "用户名和密码不能为空")
	}

	userRepo := repository.NewUserRepository(l.svcCtx.Repository)
	user, err := userRepo.FindByUsername(l.ctx, req.Username)
	if err != nil {
		// 用户不存在或查询异常统一为未授权，避免枚举用户名。
		if errors.Is(errors.Unwrap(err), model.ErrNotFound) || errors.Is(err, model.ErrNotFound) {
			// 记录登录失败日志
			l.recordLoginLog(0, req.Username, httpReq, "用户名或密码错误", false)
			return nil, errs.New(errs.CodeUnauthorized, "用户名或密码错误")
		}
		// 记录登录失败日志
		l.recordLoginLog(0, req.Username, httpReq, "查询用户失败", false)
		return nil, errs.Wrap(errs.CodeInternalError, "查询用户失败", err)
	}

	if user.Status != 1 {
		// 记录登录失败日志
		l.recordLoginLog(user.Id, user.Username, httpReq, "账号已被禁用", false)
		return nil, errs.New(errs.CodeForbidden, "账号已被禁用")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		// 记录登录失败日志
		l.recordLoginLog(user.Id, user.Username, httpReq, "用户名或密码错误", false)
		return nil, errs.New(errs.CodeUnauthorized, "用户名或密码错误")
	}

	accessToken, err := jwthelper.GenerateToken(
		l.svcCtx.Config.JWT.AccessSecret,
		l.svcCtx.Config.JWT.Issuer,
		l.svcCtx.Config.JWT.AccessExpire,
		user.Id,
		user.Username,
		false,
	)
	if err != nil {
		// 记录登录失败日志
		l.recordLoginLog(user.Id, user.Username, httpReq, "生成访问令牌失败", false)
		return nil, errs.Wrap(errs.CodeInternalError, "生成访问令牌失败", err)
	}

	refreshToken, err := jwthelper.GenerateToken(
		l.svcCtx.Config.JWT.RefreshSecret,
		l.svcCtx.Config.JWT.Issuer,
		l.svcCtx.Config.JWT.RefreshExpire,
		user.Id,
		user.Username,
		true,
	)
	if err != nil {
		// 记录登录失败日志
		l.recordLoginLog(user.Id, user.Username, httpReq, "生成刷新令牌失败", false)
		return nil, errs.Wrap(errs.CodeInternalError, "生成刷新令牌失败", err)
	}

	// 记录登录成功日志
	l.recordLoginLog(user.Id, user.Username, httpReq, "登录成功", true)

	return &types.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// recordLoginLog 记录登录日志（异步）
func (l *LoginLogic) recordLoginLog(userId uint64, username string, httpReq *http.Request, message string, success bool) {
	if httpReq == nil {
		l.Errorf("记录登录日志失败: httpReq 为 nil, userId=%d, username=%s", userId, username)
		return
	}

	// 获取 IP 地址
	ip := l.getClientIP(httpReq)

	// 获取 User-Agent
	userAgent := httpReq.UserAgent()

	// 解析浏览器和操作系统
	browser, os := useragent.ParseUserAgent(userAgent)

	// 登录状态：0失败 1成功
	status := int64(0)
	if success {
		status = 1
	}

	// 构建登录日志
	now := time.Now().Unix()
	loginLog := &model.AdminLoginLog{
		UserId:    userId,
		Username:  username,
		IpAddress: ip,
		Location:  "", // 可以通过 IP 解析地理位置，这里暂时留空
		Browser:   browser,
		Os:        os,
		UserAgent: userAgent,
		Status:    status,
		Message:   message,
		LoginAt:   now,
		LogoutAt:  0,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: 0,
	}

	// 记录调试信息
	l.Infof("准备记录登录日志: userId=%d, username=%s, status=%d, message=%s, ip=%s", userId, username, status, message, ip)

	// 异步写入日志（使用 goroutine，避免阻塞登录流程）
	go func() {
		defer func() {
			if r := recover(); r != nil {
				l.Errorf("记录登录日志时发生 panic: %v, userId=%d, username=%s", r, userId, username)
			}
		}()

		loginLogRepo := repository.NewLoginLogRepository(l.svcCtx.Repository)
		if err := loginLogRepo.Create(context.Background(), loginLog); err != nil {
			l.Errorf("记录登录日志失败: userId=%d, username=%s, status=%d, message=%s, error: %v", userId, username, status, message, err)
		} else {
			l.Infof("成功记录登录日志: userId=%d, username=%s, status=%d, message=%s", userId, username, status, message)
		}
	}()
}

// getClientIP 获取客户端 IP 地址
func (l *LoginLogic) getClientIP(r *http.Request) string {
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
