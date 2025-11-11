/**
 * @Author: zjj
 * @Date: 2025/11/6
 * @Desc:
**/

package actor

import "context"

// Message Actor消息
type Message struct {
	Context   context.Context // 上下文
	SessionId string          // 会话Id(玩家消息使用)
	MsgId     uint16          // 消息Id
	Data      []byte          // 消息数据

	// 回调相关
	ReplyTo   chan *Message // 回复通道(用于跨Actor同步调用)
	RequestId string        // 请求Id(用于异步回调)
}
