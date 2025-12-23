package gateway

import (
	"postapocgame/server/internal/database"
	"postapocgame/server/service/gameserver/internel/iface"
)

// TokenGeneratorAdapter 登录 Token 生成器
type TokenGeneratorAdapter struct{}

// NewTokenGenerator 创建 Token 生成器
func NewTokenGenerator() iface.TokenGenerator {
	return &TokenGeneratorAdapter{}
}

// Generate 生成 Token
func (g *TokenGeneratorAdapter) Generate(accountID uint64) string {
	return database.GenerateToken(uint(accountID))
}
