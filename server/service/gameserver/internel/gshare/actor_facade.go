package gshare

import (
	"postapocgame/server/internal/actor"
	"postapocgame/server/pkg/customerr"
	"sync"
)

// IActorFacade Actor门面接口
type IActorFacade interface {
	RegisterHandler(msgId uint16, f actor.HandlerMessageFunc)
	SendMessageAsync(key string, message actor.IActorMessage) error
	RemoveActor(key string) error
}

var (
	actorFacade        IActorFacade
	dungeonActorFacade IDungeonActorFacade
	facadeMu           sync.RWMutex
)

// IDungeonActorFacade DungeonActor门面接口（仅用于 GameServer 内部 PlayerActor ↔ DungeonActor 协作）
// 注意：这里只暴露 Actor 级别的发送/注册能力，不依赖具体的 dungeonactor 包，避免循环依赖。
type IDungeonActorFacade interface {
	RegisterHandler(msgId uint16, f actor.HandlerMessageFunc)
	SendMessageAsync(key string, message actor.IActorMessage) error
}

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

// SetDungeonActorFacade 设置DungeonActor门面（线程安全）
func SetDungeonActorFacade(facade IDungeonActorFacade) {
	facadeMu.Lock()
	defer facadeMu.Unlock()
	dungeonActorFacade = facade
}

// GetDungeonActorFacade 获取DungeonActor门面（线程安全）
func GetDungeonActorFacade() IDungeonActorFacade {
	facadeMu.RLock()
	defer facadeMu.RUnlock()
	return dungeonActorFacade
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
		return customerr.NewError("actor facade not initialized")
	}
	return facade.SendMessageAsync(key, message)
}

// RemoveActor 移除Actor（便捷方法）
func RemoveActor(key string) error {
	facade := GetActorFacade()
	if facade == nil {
		return customerr.NewError("actor facade not initialized")
	}
	return facade.RemoveActor(key)
}

// SendDungeonMessageAsync 发送异步消息到 DungeonActor（便捷方法）
// 约定：对于 ModeSingle 的 DungeonActor，key 一般固定为 "global"；如后续扩展多 Actor，可根据 session 决定 key。
func SendDungeonMessageAsync(key string, message actor.IActorMessage) error {
	facadeMu.RLock()
	facade := dungeonActorFacade
	facadeMu.RUnlock()
	if facade == nil {
		return customerr.NewError("dungeon actor facade not initialized")
	}
	return facade.SendMessageAsync(key, message)
}
