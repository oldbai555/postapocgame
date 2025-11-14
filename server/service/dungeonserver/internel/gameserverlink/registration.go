package gameserverlink

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
)

var (
	// 用于确保协议注册只执行一次
	protocolRegistered bool
	dungeonSrvType     uint8
	protocolProvider   func() []uint16
)

// SetDungeonSrvType 设置DungeonServer类型
func SetDungeonSrvType(srvType uint8) {
	dungeonSrvType = srvType
	log.Infof("DungeonServer srvType set to: %d", srvType)
}

// RegisterProtocolProvider 注册协议提供者
func RegisterProtocolProvider(fn func() []uint16) {
	protocolProvider = fn
}

// RegisterProtocolsToGameServer 向GameServer注册协议
// srvType: DungeonServer类型
// commonProtocols: 通用协议列表(多个DungeonServer共享)
// uniqueProtocols: 独有协议列表(仅此srvType独有)
func RegisterProtocolsToGameServer(ctx context.Context, srvType uint8, commonProtocols, uniqueProtocols []uint16) error {
	// 构造协议信息
	protocolInfos := make([]*protocol.ProtocolInfo, 0, len(commonProtocols)+len(uniqueProtocols))

	// 添加通用协议
	for _, protoId := range commonProtocols {
		protocolInfos = append(protocolInfos, &protocol.ProtocolInfo{
			ProtoId:  uint32(protoId),
			IsCommon: true,
		})
	}

	// 添加独有协议
	for _, protoId := range uniqueProtocols {
		protocolInfos = append(protocolInfos, &protocol.ProtocolInfo{
			ProtoId:  uint32(protoId),
			IsCommon: false,
		})
	}

	// 构造注册请求
	req := &protocol.D2GRegisterProtocolsReq{
		SrvType:   uint32(srvType),
		Protocols: protocolInfos,
	}

	reqData, err := proto.Marshal(req)
	if err != nil {
		log.Errorf("marshal register protocols request failed: %v", err)
		return customerr.Wrap(err)
	}

	// 发送RPC请求到GameServer
	err = CallGameServer(ctx, "", uint16(protocol.D2GRpcProtocol_D2GRegisterProtocols), reqData)
	if err != nil {
		log.Errorf("call GameServer to register protocols failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("registered %d protocols to GameServer: srvType=%d, common=%d, unique=%d",
		len(protocolInfos), srvType, len(commonProtocols), len(uniqueProtocols))

	return nil
}

// UnregisterProtocolsFromGameServer 从GameServer注销协议
func UnregisterProtocolsFromGameServer(ctx context.Context, srvType uint8) error {
	req := &protocol.D2GUnregisterProtocolsReq{
		SrvType: uint32(srvType),
	}

	reqData, err := proto.Marshal(req)
	if err != nil {
		log.Errorf("marshal unregister protocols request failed: %v", err)
		return customerr.Wrap(err)
	}

	// 发送RPC请求到GameServer
	err = CallGameServer(ctx, "", uint16(protocol.D2GRpcProtocol_D2GUnregisterProtocols), reqData)
	if err != nil {
		log.Errorf("call GameServer to unregister protocols failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("unregistered protocols from GameServer: srvType=%d", srvType)
	return nil
}

// TryRegisterProtocols 尝试注册协议到GameServer (只会执行一次)
func TryRegisterProtocols(ctx context.Context) error {
	// 如果已经注册过,直接返回
	if protocolRegistered {
		return nil
	}

	// 检查是否有可用的GameServer连接
	_, ok := GetMessageSender().GetFirstGameServer()
	if !ok {
		log.Warnf("no GameServer connection available, skip protocol registration")
		return nil
	}

	log.Infof("registering protocols to GameServer, srvType=%d", dungeonSrvType)

	if protocolProvider == nil {
		log.Warnf("protocol provider not registered, skip protocol registration")
		return nil
	}

	// 获取协议列表
	allProtocols := protocolProvider()
	// TODO: 这里需要根据实际情况配置通用协议和独有协议
	// 当前示例:所有协议都注册为通用协议
	commonProtocols := allProtocols
	uniqueProtocols := make([]uint16, 0)

	// 注册协议到GameServer
	if err := RegisterProtocolsToGameServer(ctx, dungeonSrvType, commonProtocols, uniqueProtocols); err != nil {
		log.Errorf("register protocols to GameServer failed: %v", err)
		return err
	}

	protocolRegistered = true
	log.Infof("successfully registered protocols to GameServer")

	return nil
}

// UnregisterProtocols 注销协议
func UnregisterProtocols(ctx context.Context) error {
	if !protocolRegistered {
		return nil
	}

	log.Infof("unregistering protocols from GameServer, srvType=%d", dungeonSrvType)

	if err := UnregisterProtocolsFromGameServer(ctx, dungeonSrvType); err != nil {
		log.Errorf("unregister protocols from GameServer failed: %v", err)
		return err
	}

	protocolRegistered = false
	log.Infof("successfully unregistered protocols from GameServer")

	return nil
}
