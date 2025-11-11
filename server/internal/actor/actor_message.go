/**
 * @Author: zjj
 * @Date: 2025/11/11
 * @Desc:
**/

package actor

import "context"

var _ IActorMessage = (*BaseMessage)(nil)

type BaseMessage struct {
	Context context.Context // 上下文数据
	MsgId   uint16          // 消息ID
	Data    []byte          // 消息数据
}

func NewBaseMessage(ctx context.Context, msgId uint16, data []byte) IActorMessage {
	return &BaseMessage{
		Context: ctx,
		MsgId:   msgId,
		Data:    data,
	}
}

func (m *BaseMessage) GetMsgId() uint16 {
	return m.MsgId
}

func (m *BaseMessage) GetData() []byte {
	return m.Data
}

func (m *BaseMessage) GetContext() context.Context {
	return m.Context
}
