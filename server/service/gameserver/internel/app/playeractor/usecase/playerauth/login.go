package playerauth

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
	"strings"
)

// LoginInput 登录入参
type LoginInput struct {
	Username string
	Password string
}

// LoginResult 登录结果
type LoginResult struct {
	Success   bool
	Message   string
	Token     string
	AccountID uint64
}

// LoginUseCase 登录用例
type LoginUseCase struct {
	accountRepo   repository.AccountRepository
	tokenProvider interfaces.TokenGenerator
}

// NewLoginUseCase 创建登录用例
func NewLoginUseCase(repo repository.AccountRepository, tokenProvider interfaces.TokenGenerator) *LoginUseCase {
	return &LoginUseCase{
		accountRepo:   repo,
		tokenProvider: tokenProvider,
	}
}

// Execute 执行登录
func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginResult, error) {
	username := strings.TrimSpace(input.Username)
	password := strings.TrimSpace(input.Password)
	if username == "" || password == "" {
		return &LoginResult{
			Success: false,
			Message: "用户名或密码错误",
		}, nil
	}

	account, err := uc.accountRepo.GetAccountByUsername(ctx, username)
	if err != nil {
		if err == repository.ErrAccountNotFound {
			return &LoginResult{
				Success: false,
				Message: "用户名或密码错误",
			}, nil
		}
		return nil, err
	}
	if account == nil {
		return &LoginResult{
			Success: false,
			Message: "用户名或密码错误",
		}, nil
	}

	if !account.CheckPassword(password) {
		return &LoginResult{
			Success: false,
			Message: "用户名或密码错误",
		}, nil
	}

	token := uc.tokenProvider.Generate(account.ID)
	return &LoginResult{
		Success:   true,
		Message:   "登录成功",
		Token:     token,
		AccountID: account.ID,
	}, nil
}
