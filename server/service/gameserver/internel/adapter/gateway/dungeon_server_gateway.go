package gateway

import (
	"context"
	dungeonserverlink2 "postapocgame/server/service/gameserver/internel/infrastructure/dungeonserverlink"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// DungeonServerGatewayImpl DungeonServer RPC 实现
type DungeonServerGatewayImpl struct{}

// NewDungeonServerGateway 创建 DungeonServer Gateway
func NewDungeonServerGateway() interfaces.DungeonServerGateway {
	return &DungeonServerGatewayImpl{}
}

// AsyncCall 异步调用 DungeonServer
func (g *DungeonServerGatewayImpl) AsyncCall(ctx context.Context, srvType uint8, sessionId string, msgId uint16, data []byte) error {
	return dungeonserverlink2.AsyncCall(ctx, srvType, sessionId, msgId, data)
}

// RegisterRPCHandler 注册 RPC 处理器
func (g *DungeonServerGatewayImpl) RegisterRPCHandler(msgId uint16, handler func(ctx context.Context, sessionId string, data []byte) error) {
	dungeonserverlink2.RegisterRPCHandler(msgId, handler)
}

// IsDungeonProtocol 判断是否是 DungeonServer 协议
func (g *DungeonServerGatewayImpl) IsDungeonProtocol(protoId uint16) bool {
	return dungeonserverlink2.GetProtocolManager().IsDungeonProtocol(protoId)
}

// GetSrvTypeForProtocol 获取协议对应的 srvType
func (g *DungeonServerGatewayImpl) GetSrvTypeForProtocol(protoId uint16) (uint8, interfaces.ProtocolType, bool) {
	srvType, protocolType, ok := dungeonserverlink2.GetProtocolManager().GetSrvTypeForProtocol(protoId)
	return srvType, interfaces.ProtocolType(protocolType), ok
}

// RegisterProtocols 注册 DungeonServer 协议
func (g *DungeonServerGatewayImpl) RegisterProtocols(srvType uint8, protocols []interfaces.ProtocolInfo) error {
	converted := make([]struct {
		ProtoId  uint16
		IsCommon bool
	}, len(protocols))
	for i, info := range protocols {
		converted[i].ProtoId = info.ProtoID
		converted[i].IsCommon = info.IsCommon
	}
	return dungeonserverlink2.GetProtocolManager().RegisterProtocols(srvType, converted)
}

// UnregisterProtocols 注销 DungeonServer 协议
func (g *DungeonServerGatewayImpl) UnregisterProtocols(srvType uint8) error {
	return dungeonserverlink2.GetProtocolManager().UnregisterProtocols(srvType)
}
