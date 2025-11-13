package dungeonserverlink

import (
	"context"
	"postapocgame/server/internal"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
)

// handleRegisterProtocols 处理DungeonServer注册协议的RPC请求
func handleRegisterProtocols(ctx context.Context, sessionId string, data []byte) error {
	var req protocol.D2GRegisterProtocolsReq
	if err := internal.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal register protocols request failed: %v", err)
		return customerr.Wrap(err)
	}

	srvType := uint8(req.SrvType)
	log.Infof("received protocol registration from DungeonServer: srvType=%d, protocols=%d", srvType, len(req.Protocols))

	// 转换协议信息
	protocols := make([]struct {
		ProtoId  uint16
		IsCommon bool
	}, len(req.Protocols))

	for i, proto := range req.Protocols {
		protocols[i].ProtoId = uint16(proto.ProtoId)
		protocols[i].IsCommon = proto.IsCommon
		log.Debugf("  - protoId=%d, isCommon=%v", proto.ProtoId, proto.IsCommon)
	}

	// 注册到协议管理器
	if err := GetProtocolManager().RegisterProtocols(srvType, protocols); err != nil {
		log.Errorf("register protocols failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("successfully registered %d protocols for srvType=%d", len(protocols), srvType)
	return nil
}

// handleUnregisterProtocols 处理DungeonServer注销协议的RPC请求
func handleUnregisterProtocols(ctx context.Context, sessionId string, data []byte) error {
	var req protocol.D2GUnregisterProtocolsReq
	if err := internal.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal unregister protocols request failed: %v", err)
		return customerr.Wrap(err)
	}

	srvType := uint8(req.SrvType)
	log.Infof("received protocol unregistration from DungeonServer: srvType=%d", srvType)

	// 从协议管理器注销
	if err := GetProtocolManager().UnregisterProtocols(srvType); err != nil {
		log.Errorf("unregister protocols failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("successfully unregistered protocols for srvType=%d", srvType)
	return nil
}

// InitProtocolRegistration 初始化协议注册相关的RPC处理器
func InitProtocolRegistration() {
	// 注册协议注册的RPC处理器
	RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GRegisterProtocols), handleRegisterProtocols)
	RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GUnregisterProtocols), handleUnregisterProtocols)

	log.Infof("protocol registration RPC handlers initialized")
}
