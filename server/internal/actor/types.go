/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package actor

type Mode = uint8

const (
	ModeBySingle    Mode = iota // 单Actor模式(所有消息串行处理)
	ModeByPerPlayer             // 每玩家一个Actor模式
)

type ActorType int //  特殊Actor类型

const (
	SatPublic ActorType = iota // 公共数据Actor(处理全局数据)
)

func IsModeByPerPlayer(mode int) bool {
	return ModeByPerPlayer == Mode(mode)
}

type SendFunc = func(sessionId string, msgId uint16, data []byte) error
type RegisterFunc = func(msgId uint16, handler MsgHandlerFunc)
type RemovePerPlayerActorFunc = func(sessionId string) error
