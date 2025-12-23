package playerauth

import (
	"context"
	"postapocgame/server/service/gameserver/internel/iface"
	"strings"
)

// RegisterInput 注册入参
type RegisterInput struct {
	Username string
	Password string
}

// RegisterResult 注册结果
type RegisterResult struct {
	Success   bool
	Message   string
	Token     string
	AccountID uint64
}

// RegisterUseCase 账号注册用例
type RegisterUseCase struct {
	accountRepo   iface.AccountRepository
	tokenProvider iface.TokenGenerator
}

// NewRegisterUseCase 创建注册用例
func NewRegisterUseCase(repo iface.AccountRepository, tokenProvider iface.TokenGenerator) *RegisterUseCase {
	return &RegisterUseCase{
		accountRepo:   repo,
		tokenProvider: tokenProvider,
	}
}

// Execute 执行注册
func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterResult, error) {
	username := strings.TrimSpace(input.Username)
	password := strings.TrimSpace(input.Password)

	if len(username) < 3 || len(username) > 32 {
		return &RegisterResult{
			Success: false,
			Message: "用户名长度必须在3-32个字符之间",
		}, nil
	}
	if len(password) < 6 {
		return &RegisterResult{
			Success: false,
			Message: "密码长度至少6个字符",
		}, nil
	}

	_, err := uc.accountRepo.GetAccountByUsername(ctx, username)
	if err == nil {
		return &RegisterResult{
			Success: false,
			Message: "用户名已存在",
		}, nil
	}
	if err != nil && err != iface.ErrAccountNotFound {
		return nil, err
	}

	account, err := uc.accountRepo.CreateAccount(ctx, username, password)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return &RegisterResult{
			Success: false,
			Message: "注册失败",
		}, nil
	}

	token := uc.tokenProvider.Generate(account.ID)
	return &RegisterResult{
		Success:   true,
		Message:   "注册成功",
		Token:     token,
		AccountID: account.ID,
	}, nil
}
