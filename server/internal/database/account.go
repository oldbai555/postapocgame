package database

import "golang.org/x/crypto/bcrypt"

// Account 账号表
type Account struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"unique;not null;size:32"`
	Password  string `gorm:"not null;size:128"` // 存储bcrypt hash
	CreatedAt int64  `gorm:"autoCreateTime"`
	UpdatedAt int64  `gorm:"autoUpdateTime"`
}

// SetPassword 对密码加密并设置
func (a *Account) SetPassword(pw string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.Password = string(hash)
	return nil
}

// CheckPassword 检查密码
func (a *Account) CheckPassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(pw)) == nil
}

// CreateAccount 注册账号
func CreateAccount(username, pw string) (*Account, error) {
	acct := &Account{Username: username}
	if err := acct.SetPassword(pw); err != nil {
		return nil, err
	}
	result := DB.Create(acct)
	return acct, result.Error
}

// GetAccountByUsername 通过用户名查找
func GetAccountByUsername(username string) (*Account, error) {
	var acct Account
	result := DB.Where("username = ?", username).First(&acct)
	if result.Error != nil {
		return nil, result.Error
	}
	return &acct, nil
}
