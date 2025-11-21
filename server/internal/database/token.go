package database

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"postapocgame/server/internal/servertime"
)

// GenerateToken 生成登录token
func GenerateToken(accountID uint) string {
	// 使用账号ID + 时间戳 + 随机数生成token
	timestamp := servertime.Now().Unix()
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)

	tokenStr := fmt.Sprintf("%d_%d_%s", accountID, timestamp, hex.EncodeToString(randomBytes))
	return tokenStr
}

// ValidateToken 验证token格式（简单验证，实际可以加入过期时间等）
func ValidateToken(token string) bool {
	// 简单验证token格式
	return len(token) > 0
}
