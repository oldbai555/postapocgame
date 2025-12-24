package jwt

import (
	"time"

	jwtv4 "github.com/golang-jwt/jwt/v4"
)

// Claims 自定义 JWT 声明，包含用户基础信息与是否为刷新令牌。
type Claims struct {
	UserID    uint64 `json:"uid"`
	Username  string `json:"uname"`
	IsRefresh bool   `json:"isRefresh"`
	jwtv4.RegisteredClaims
}

func GenerateToken(secret, issuer string, expireSeconds int64, userID uint64, username string, isRefresh bool) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Username:  username,
		IsRefresh: isRefresh,
		RegisteredClaims: jwtv4.RegisteredClaims{
			Issuer:    issuer,
			IssuedAt:  jwtv4.NewNumericDate(now),
			ExpiresAt: jwtv4.NewNumericDate(now.Add(time.Duration(expireSeconds) * time.Second)),
		},
	}

	token := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
