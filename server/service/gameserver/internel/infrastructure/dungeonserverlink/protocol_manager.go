package dungeonserverlink

import (
	"fmt"
	"sync"

	"postapocgame/server/pkg/log"
)

// ProtocolType 协议类型
type ProtocolType uint8

const (
	ProtocolTypeUnknown ProtocolType = 0 // 未知协议
	ProtocolTypeCommon  ProtocolType = 1 // 通用协议(多个DungeonServer共享)
	ProtocolTypeUnique  ProtocolType = 2 // 独有协议(特定srvType的DungeonServer)
)

// ProtocolRoute 协议路由信息
type ProtocolRoute struct {
	ProtoId  uint16       // 协议ID
	Type     ProtocolType // 协议类型
	SrvType  uint8        // 如果是独有协议,指定srvType
	IsCommon bool         // 是否为通用协议
}

// ProtocolManager 协议管理器 - 管理DungeonServer的协议注册信息
type ProtocolManager struct {
	mu sync.RWMutex

	// 协议路由表: protoId -> ProtocolRoute
	routes map[uint16]*ProtocolRoute

	// srvType的协议列表: srvType -> []protoId
	srvTypeProtocols map[uint8][]uint16
}

var (
	protocolManager     *ProtocolManager
	protocolManagerOnce sync.Once
)

// GetProtocolManager 获取协议管理器单例
func GetProtocolManager() *ProtocolManager {
	protocolManagerOnce.Do(func() {
		protocolManager = &ProtocolManager{
			routes:           make(map[uint16]*ProtocolRoute),
			srvTypeProtocols: make(map[uint8][]uint16),
		}
	})
	return protocolManager
}

// RegisterProtocols 注册DungeonServer的协议
func (pm *ProtocolManager) RegisterProtocols(srvType uint8, protocols []struct {
	ProtoId  uint16
	IsCommon bool
}) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 检查srvType是否已经注册
	if _, exists := pm.srvTypeProtocols[srvType]; exists {
		return fmt.Errorf("srvType %d already registered", srvType)
	}

	registeredProtos := make([]uint16, 0, len(protocols))

	for _, proto := range protocols {
		protoId := proto.ProtoId

		// 检查协议是否已存在
		if existing, exists := pm.routes[protoId]; exists {
			// 如果已存在,检查是否一致
			if existing.IsCommon != proto.IsCommon {
				log.Warnf("protocol %d conflict: existing isCommon=%v, new isCommon=%v",
					protoId, existing.IsCommon, proto.IsCommon)
			}

			// 如果是通用协议,不需要重复注册
			if proto.IsCommon {
				continue
			}

			// 如果是独有协议,检查srvType是否冲突
			if !existing.IsCommon {
				return fmt.Errorf("protocol %d already registered by srvType %d", protoId, existing.SrvType)
			}
		}

		// 注册新协议
		route := &ProtocolRoute{
			ProtoId:  protoId,
			IsCommon: proto.IsCommon,
			SrvType:  srvType,
		}

		if proto.IsCommon {
			route.Type = ProtocolTypeCommon
		} else {
			route.Type = ProtocolTypeUnique
		}

		pm.routes[protoId] = route
		registeredProtos = append(registeredProtos, protoId)

		log.Debugf("registered protocol: protoId=%d, srvType=%d, isCommon=%v", protoId, srvType, proto.IsCommon)
	}

	// 记录srvType的协议列表
	pm.srvTypeProtocols[srvType] = registeredProtos

	log.Infof("registered %d protocols for srvType=%d", len(registeredProtos), srvType)
	return nil
}

// UnregisterProtocols 注销DungeonServer的协议
func (pm *ProtocolManager) UnregisterProtocols(srvType uint8) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	protoIds, exists := pm.srvTypeProtocols[srvType]
	if !exists {
		return fmt.Errorf("srvType %d not registered", srvType)
	}

	// 删除协议路由
	for _, protoId := range protoIds {
		if route, exists := pm.routes[protoId]; exists {
			// 只删除属于该srvType的协议
			if route.SrvType == srvType {
				delete(pm.routes, protoId)
				log.Debugf("unregistered protocol: protoId=%d, srvType=%d", protoId, srvType)
			}
		}
	}

	// 删除srvType记录
	delete(pm.srvTypeProtocols, srvType)

	log.Infof("unregistered %d protocols for srvType=%d", len(protoIds), srvType)
	return nil
}

// GetProtocolRoute 获取协议路由信息
func (pm *ProtocolManager) GetProtocolRoute(protoId uint16) (*ProtocolRoute, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	route, exists := pm.routes[protoId]
	return route, exists
}

// IsDungeonProtocol 检查协议是否需要转发到DungeonServer
func (pm *ProtocolManager) IsDungeonProtocol(protoId uint16) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	_, exists := pm.routes[protoId]
	return exists
}

// GetSrvTypeForProtocol 获取协议对应的srvType
// 对于通用协议,需要根据角色所在的DungeonServer来确定
// 对于独有协议,返回注册时指定的srvType
func (pm *ProtocolManager) GetSrvTypeForProtocol(protoId uint16) (uint8, ProtocolType, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	route, exists := pm.routes[protoId]
	if !exists {
		return 0, ProtocolTypeUnknown, false
	}

	return route.SrvType, route.Type, true
}

// GetRegisteredSrvTypes 获取所有已注册的srvType列表
func (pm *ProtocolManager) GetRegisteredSrvTypes() []uint8 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	srvTypes := make([]uint8, 0, len(pm.srvTypeProtocols))
	for srvType := range pm.srvTypeProtocols {
		srvTypes = append(srvTypes, srvType)
	}
	return srvTypes
}

// Clear 清空所有协议注册信息
func (pm *ProtocolManager) Clear() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.routes = make(map[uint16]*ProtocolRoute)
	pm.srvTypeProtocols = make(map[uint8][]uint16)

	log.Infof("protocol manager cleared")
}
