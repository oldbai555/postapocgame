/**
 * @Author: zjj
 * @Date: 2025/11/11
 * @Desc:
**/

package base

import "postapocgame/server/internal/actor"

type SessionMessage struct {
	actor.BaseMessage
	SessionId string
}

func NewSessionMessage() *SessionMessage {
	return &SessionMessage{}
}
