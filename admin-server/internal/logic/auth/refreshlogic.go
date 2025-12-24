// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"

	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshLogic {
	return &RefreshLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshLogic) Refresh(req *types.RefreshReq) (resp *types.TokenPair, err error) {
	if req == nil || req.RefreshToken == "" {
		return nil, errs.New(errs.CodeBadRequest, "刷新令牌不能为空")
	}

	claims, err := jwthelper.ParseToken(req.RefreshToken, l.svcCtx.Config.JWT.RefreshSecret)
	if err != nil || !claims.IsRefresh {
		return nil, errs.New(errs.CodeUnauthorized, "刷新令牌无效或已过期")
	}

	accessToken, err := jwthelper.GenerateToken(
		l.svcCtx.Config.JWT.AccessSecret,
		l.svcCtx.Config.JWT.Issuer,
		l.svcCtx.Config.JWT.AccessExpire,
		claims.UserID,
		claims.Username,
		false,
	)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "生成访问令牌失败", err)
	}

	refreshToken, err := jwthelper.GenerateToken(
		l.svcCtx.Config.JWT.RefreshSecret,
		l.svcCtx.Config.JWT.Issuer,
		l.svcCtx.Config.JWT.RefreshExpire,
		claims.UserID,
		claims.Username,
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
