package pubilcactor

import (
	"postapocgame/server/internal/actor"
	"postapocgame/server/service/gameserver/internel/gshare"
)

func NewActorSystem() actor.IActorSystem {
	var a actor.BaseActorMsgHandler
	system := actor.NewActorSystem(actor.ModeBySingle, &a)
	gshare.PublicSendFunc = system.Send
	gshare.PublicRegisterFunc = a.Register
	return system
}
