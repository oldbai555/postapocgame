package gateway

import (
	"context"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/dungeonactor"
	"postapocgame/server/service/gameserver/internel/iface"
)

// DungeonServerGatewayImpl DungeonServer Gateway 实现
// 将调用路由到 GameServer 进程内的 DungeonActor
type DungeonServerGatewayImpl struct{}

// NewDungeonServerGateway 创建 DungeonServer Gateway
func NewDungeonServerGateway() iface.DungeonServerGateway {
	return &DungeonServerGatewayImpl{}
}

// AsyncCall 调用 DungeonActor
func (g *DungeonServerGatewayImpl) AsyncCall(ctx context.Context, sessionId string, msgId uint16, data []byte) error {
	da := dungeonactor.GetDungeonActor()
	if da == nil {
		return customerr.NewError("dungeon actor not initialized")
	}
	return da.AsyncCall(ctx, sessionId, msgId, data)
}
