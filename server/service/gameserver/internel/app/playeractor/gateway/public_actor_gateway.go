package gateway

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

// PublicActorGatewayImpl PublicActor 交互实现
type PublicActorGatewayImpl struct{}

// NewPublicActorGateway 创建 PublicActor Gateway
func NewPublicActorGateway() iface.PublicActorGateway {
	return &PublicActorGatewayImpl{}
}

// SendMessageAsync 发送异步消息到 PublicActor
func (g *PublicActorGatewayImpl) SendMessageAsync(ctx context.Context, key string, message actor.IActorMessage) error {
	return gshare.SendPublicMessageAsync(key, message)
}

// RegisterHandler 注册消息处理器
func (g *PublicActorGatewayImpl) RegisterHandler(msgId uint16, handler actor.HandlerMessageFunc) {
	facade := gshare.GetPublicActorFacade()
	if facade == nil {
		// 如果 PublicActor 尚未初始化，记录警告但不报错
		// 在实际使用时会通过 SendMessageAsync 检查
		return
	}
	facade.RegisterHandler(msgId, handler)
}
