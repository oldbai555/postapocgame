package event

import (
	"postapocgame/server/pkg/tool"
	"time"
)

// Event 事件（增强版）
type Event struct {
	Type      Type                   // 事件类型
	Data      []interface{}          // 事件数据
	Source    string                 // 事件源（actor ID, server ID 等）
	Timestamp int64                  // 事件时间戳（毫秒）
	Metadata  map[string]interface{} // 额外元数据
}

// NewEvent 创建新事件
func NewEvent(eventType Type, data ...interface{}) *Event {
	return &Event{
		Type:      eventType,
		Data:      data,
		Source:    tool.GetCaller(1),
		Timestamp: time.Now().UnixMilli(),
		Metadata:  make(map[string]interface{}),
	}
}

// SetMetadata 设置元数据
func (e *Event) SetMetadata(key string, value interface{}) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
}

// GetMetadata 获取元数据
func (e *Event) GetMetadata(key string) (interface{}, bool) {
	if e.Metadata == nil {
		return nil, false
	}
	val, ok := e.Metadata[key]
	return val, ok
}
