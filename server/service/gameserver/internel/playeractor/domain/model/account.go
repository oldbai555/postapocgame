package model

import "golang.org/x/crypto/bcrypt"

// Account 领域层账号实体
type Account struct {
	ID           uint64
	Username     string
	passwordHash string
}

// NewAccount 创建账号实体
func NewAccount(id uint64, username string, passwordHash string) *Account {
	return &Account{
		ID:           id,
		Username:     username,
		passwordHash: passwordHash,
	}
}

// CheckPassword 校验密码
func (a *Account) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.passwordHash), []byte(password)) == nil
}
