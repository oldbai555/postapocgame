/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package iface

import (
	"context"
)

type IGameServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type ISession interface {
	SetRoleId(roleId uint64)
	GetSessionId() string
	GetRoleId() uint64
	SetAccountID(accountID uint)
	GetAccountID() uint
	SetToken(token string)
	GetToken() string
}

type IDungeonRPC interface {
	AsyncCall(ctx context.Context, srvType uint8, msgId uint16, data []byte) error
	// Connect 连接到DungeonServer
	Connect(ctx context.Context, srvType uint8, addr string) error
	// Close 关闭连接
	Close() error
}
