package interfaces

// TokenGenerator 登录 Token 生成器
type TokenGenerator interface {
	Generate(accountID uint64) string
}
