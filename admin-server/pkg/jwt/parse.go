package jwt

import (
	"fmt"

	jwtv4 "github.com/golang-jwt/jwt/v4"
)

// ParseToken 解析并验证 JWT，返回自定义 Claims。
func ParseToken(tokenStr, secret string) (*Claims, error) {
	if tokenStr == "" {
		return nil, fmt.Errorf("empty token")
	}

	token, err := jwtv4.ParseWithClaims(tokenStr, &Claims{}, func(token *jwtv4.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv4.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
