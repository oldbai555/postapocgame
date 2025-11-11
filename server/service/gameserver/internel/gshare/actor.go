/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package gshare

import "postapocgame/server/internal/actor"

var (
	PublicSendFunc         actor.SendFunc
	PlayerRoleSendFunc     actor.SendFunc
	PublicRegisterFunc     actor.RegisterFunc
	PlayerRoleRegisterFunc actor.RegisterFunc

	PlayerRoleRemovePerPlayerActorFunc actor.RemovePerPlayerActorFunc
)
