package event

import (
	"postapocgame/server/pkg/log"
)

// GlobalPublish 发布全局事件（广播到所有 actor）
func GlobalPublish(event *Event) {
	initEventSystem()

	if event.Source == "" {
		event.Source = "global"
	}

	// 广播到所有 actor 的邮箱
	actorRegistry.Broadcast(event)

	log.Debugf("[EventSystem] Global event broadcasted: type=%d, source=%s, actors=%d",
		event.Type, event.Source, actorRegistry.GetActorCount())
}

// GlobalPublishToActors 发布全局事件到指定的 actors
func GlobalPublishToActors(event *Event, actorIDs []string) {
	initEventSystem()

	if event.Source == "" {
		event.Source = "global"
	}

	actorRegistry.BroadcastToActors(event, actorIDs)

	log.Debugf("[EventSystem] Global event sent to actors: type=%d, count=%d",
		event.Type, len(actorIDs))
}

// SendToActor 发送事件到指定 actor
func SendToActor(actorID string, event *Event) bool {
	initEventSystem()

	success := actorRegistry.SendToActor(actorID, event)

	if !success {
		log.Warnf("[EventSystem] Failed to send event to actor: actorID=%s, type=%d",
			actorID, event.Type)
	}

	return success
}

// GetActorCount 获取当前 actor 数量
func GetActorCount() int {
	initEventSystem()
	return actorRegistry.GetActorCount()
}

// HasActor 检查 actor 是否存在
func HasActor(actorID string) bool {
	initEventSystem()
	return actorRegistry.HasActor(actorID)
}

// GetAllActorIDs 获取所有 actor IDs
func GetAllActorIDs() []string {
	initEventSystem()
	return actorRegistry.GetAllActorIDs()
}

// CloseEventSystem 关闭事件系统（用于测试或服务器关闭）
func CloseEventSystem() {
	if actorRegistry != nil {
		actorRegistry.Close()
	}
}
