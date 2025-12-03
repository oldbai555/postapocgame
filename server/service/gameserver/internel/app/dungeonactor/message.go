package dungeonactor

import "context"

// DungeonActorMessage 封装发送给 DungeonActor 的内部消息
type DungeonActorMessage struct {
	ctx       context.Context
	msgId     uint16
	sessionId string
	data      []byte
}

func (m *DungeonActorMessage) GetMsgId() uint16 {
	return m.msgId
}

func (m *DungeonActorMessage) GetData() []byte {
	return m.data
}

func (m *DungeonActorMessage) GetContext() context.Context {
	return m.ctx
}
