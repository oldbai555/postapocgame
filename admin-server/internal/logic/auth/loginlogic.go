// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"
	"errors"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"

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

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.TokenPair, err error) {
	if req == nil || req.Username == "" || req.Password == "" {
		return nil, errs.New(errs.CodeBadRequest, "用户名和密码不能为空")
	}

	userRepo := repository.NewUserRepository(l.svcCtx.Repository)
	user, err := userRepo.FindByUsername(l.ctx, req.Username)
	if err != nil {
		// 用户不存在或查询异常统一为未授权，避免枚举用户名。
		if errors.Is(errors.Unwrap(err), model.ErrNotFound) || errors.Is(err, model.ErrNotFound) {
			return nil, errs.New(errs.CodeUnauthorized, "用户名或密码错误")
		}
		return nil, errs.Wrap(errs.CodeInternalError, "查询用户失败", err)
	}

	if user.Status != 1 {
		return nil, errs.New(errs.CodeForbidden, "账号已被禁用")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
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
		return nil, errs.Wrap(errs.CodeInternalError, "生成刷新令牌失败", err)
	}

	return &types.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
