/**
 * @Author: zjj
 * @Date: 2025/11/11
 * @Desc:
**/

package dshare

import "postapocgame/server/internal/actor"

var (
	RegisterHandler  func(msgId uint16, f actor.HandlerMessageFunc)
	SendMessageAsync func(key string, message actor.IActorMessage) error
)
