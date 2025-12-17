package gateway

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/model"
	"postapocgame/server/service/gameserver/internel/iface"

	"gorm.io/gorm"

	"postapocgame/server/internal/database"
)

// AccountGateway 账号数据访问实现
type AccountGateway struct{}

// NewAccountGateway 创建账号 Gateway
func NewAccountGateway() iface.AccountRepository {
	return &AccountGateway{}
}

// CreateAccount 创建账号
func (g *AccountGateway) CreateAccount(_ context.Context, username, password string) (*model.Account, error) {
	acct, err := database.CreateAccount(username, password)
	if err != nil {
		return nil, err
	}
	return convertAccount(acct), nil
}

// GetAccountByUsername 通过用户名查找账号
func (g *AccountGateway) GetAccountByUsername(_ context.Context, username string) (*model.Account, error) {
	acct, err := database.GetAccountByUsername(username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iface.ErrAccountNotFound
		}
		return nil, err
	}
	return convertAccount(acct), nil
}

func convertAccount(acct *database.Account) *model.Account {
	if acct == nil {
		return nil
	}
	return model.NewAccount(uint64(acct.ID), acct.Username, acct.Password)
}
