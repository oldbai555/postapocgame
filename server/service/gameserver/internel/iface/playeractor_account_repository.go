package iface

import (
	"context"
	"errors"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/model"
)

var (
	// ErrAccountNotFound 账号不存在
	ErrAccountNotFound = errors.New("account not found")
)

// AccountRepository 账号数据访问接口
type AccountRepository interface {
	CreateAccount(ctx context.Context, username, password string) (*model.Account, error)
	GetAccountByUsername(ctx context.Context, username string) (*model.Account, error)
}
