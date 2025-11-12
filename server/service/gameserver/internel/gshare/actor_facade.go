package gshare

import (
	"fmt"
	"postapocgame/server/internal/actor"
	"sync"
)

// IActorFacade Actor门面接口
type IActorFacade interface {
	RegisterHandler(msgId uint16, f actor.HandlerMessageFunc)
	SendMessageAsync(key string, message actor.IActorMessage) error
	RemoveActor(key string) error
}

var (
	actorFacade IActorFacade
	facadeMu    sync.RWMutex
)

// SetActorFacade 设置Actor门面（线程安全）
func SetActorFacade(facade IActorFacade) {
	facadeMu.Lock()
	defer facadeMu.Unlock()
	actorFacade = facade
}

// GetActorFacade 获取Actor门面（线程安全）
func GetActorFacade() IActorFacade {
	facadeMu.RLock()
	defer facadeMu.RUnlock()
	return actorFacade
}

// RegisterHandler 注册消息处理器（便捷方法）
func RegisterHandler(msgId uint16, f actor.HandlerMessageFunc) {
	if facade := GetActorFacade(); facade != nil {
		facade.RegisterHandler(msgId, f)
	}
}

// SendMessageAsync 发送异步消息（便捷方法）
func SendMessageAsync(key string, message actor.IActorMessage) error {
	facade := GetActorFacade()
	if facade == nil {
		return fmt.Errorf("actor facade not initialized")
	}
	return facade.SendMessageAsync(key, message)
}

// RemoveActor 移除Actor（便捷方法）
func RemoveActor(key string) error {
	facade := GetActorFacade()
	if facade == nil {
		return fmt.Errorf("actor facade not initialized")
	}
	return facade.RemoveActor(key)
}
