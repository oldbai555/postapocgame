package publicactor

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/core/gshare"
)

// getPublicRole 通过 PublicActorFacade 反查当前 PublicRole 实例
func getPublicRole() *PublicRole {
	facade := gshare.GetPublicActorFacade()
	if facade == nil {
		log.Warnf("getPublicRole: public actor facade is nil")
		return nil
	}
	// 当前实现中 SetPublicActorFacade 始终注入 *PublicActor
	if pa, ok := facade.(*PublicActor); ok && pa.publicHandler != nil {
		return pa.publicHandler.publicRole
	}
	log.Warnf("getPublicRole: facade is not *PublicActor")
	return nil
}

// withPublicRole 将带 PublicRole 的业务函数适配为 actor.HandlerMessageFunc
func withPublicRole(fn func(ctx context.Context, msg actor.IActorMessage, pr *PublicRole)) actor.HandlerMessageFunc {
	return func(msg actor.IActorMessage) {
		ctx := msg.GetContext()
		publicRole := getPublicRole()
		if publicRole == nil {
			return
		}
		fn(ctx, msg, publicRole)
	}
}
