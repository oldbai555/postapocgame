/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package playeractor

import (
	"postapocgame/server/internal/actor"
	"postapocgame/server/service/gameserver/internel/gshare"
)

func NewActorSystem(mode int) actor.IActorSystem {
	var a = NewPlayerHandler()
	system := actor.NewActorSystem(actor.Mode(mode), a)
	gshare.PlayerRoleSendFunc = system.Send
	gshare.PlayerRoleRegisterFunc = a.Register
	gshare.PlayerRoleRemovePerPlayerActorFunc = system.RemovePerPlayerActor
	return system
}
