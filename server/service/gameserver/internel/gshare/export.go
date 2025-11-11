/**
 * @Author: zjj
 * @Date: 2025/11/11
 * @Desc:
**/

package gshare

import "postapocgame/server/internal/actor"

var (
	PlayerRegisterHandler  func(msgId uint16, f actor.HandlerMessageFunc)
	PlayerSendMessageAsync func(key string, message actor.IActorMessage) error
	PlayerRemoveActor      func(key string) error
)
