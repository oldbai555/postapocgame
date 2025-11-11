package event

import (
	"time"
)

// Event 事件（增强版）
type Event struct {
	Type      Type                   // 事件类型
	Data      []interface{}          // 事件数据
	Source    string                 // 事件源（actor ID, server ID 等）
	Timestamp int64                  // 事件时间戳（毫秒）
	TraceID   string                 // 追踪 ID（用于日志追踪）
	Metadata  map[string]interface{} // 额外元数据
}

// NewEvent 创建新事件
func NewEvent(eventType Type, source string, data ...interface{}) *Event {
	return &Event{
		Type:      eventType,
		Data:      data,
		Source:    source,
		Timestamp: time.Now().UnixMilli(),
		Metadata:  make(map[string]interface{}),
	}
}

// NewEventWithTrace 创建带追踪 ID 的事件
func NewEventWithTrace(eventType Type, source, traceID string, data ...interface{}) *Event {
	return &Event{
		Type:      eventType,
		Data:      data,
		Source:    source,
		Timestamp: time.Now().UnixMilli(),
		TraceID:   traceID,
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

// GetDataAt 获取指定索引的数据
func (e *Event) GetDataAt(index int) interface{} {
	if index >= 0 && index < len(e.Data) {
		return e.Data[index]
	}
	return nil
}

// GetDataCount 获取数据数量
func (e *Event) GetDataCount() int {
	return len(e.Data)
}

// Clone 克隆事件
func (e *Event) Clone() *Event {
	data := make([]interface{}, len(e.Data))
	copy(data, e.Data)

	metadata := make(map[string]interface{})
	for k, v := range e.Metadata {
		metadata[k] = v
	}

	return &Event{
		Type:      e.Type,
		Data:      data,
		Source:    e.Source,
		Timestamp: e.Timestamp,
		TraceID:   e.TraceID,
		Metadata:  metadata,
	}
}
