/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package dungeonactor

import (
	"postapocgame/server/internal/actor"
)

var SendFunc actor.SendFunc
var RegisterFunc actor.RegisterFunc

func NewActorSystem() actor.IActorSystem {
	var a = &ActorMsgHandler{}
	system := actor.NewActorSystem(actor.ModeBySingle, a)
	SendFunc = system.Send
	RegisterFunc = a.Register
	return system
}
