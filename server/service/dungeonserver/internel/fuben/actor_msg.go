/**
 * @Author: zjj
 * @Date: 2025/11/10
 * @Desc:
**/

package fuben

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/base"
	"postapocgame/server/service/dungeonserver/internel/devent"
	"postapocgame/server/service/dungeonserver/internel/dshare"
	"postapocgame/server/service/dungeonserver/internel/entity"
)

func init() {
	devent.Subscribe(devent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		dshare.RegisterHandler(protocol.RPC_EnterDungeon, func(message actor.IActorMessage) {
			msg := message.(*base.SessionMessage)
			if msg.SessionId == "" {
				return
			}
			var roleInfo protocol.RoleInfo
			err := tool.JsonUnmarshal(msg.Data, &roleInfo)
			if err != nil {
				return
			}
			roleEntity := entity.NewRoleEntity(msg.SessionId, &roleInfo)
			resp := protocol.EnterSceneResponse{
				SceneId:  1,
				RoleInfo: &roleInfo,
				PosX:     1,
				PosY:     1,
			}
			bytes, err := tool.JsonMarshal(resp)
			if err != nil {
				return
			}
			err = roleEntity.SendMessage(1, 3, bytes)
			if err != nil {
				log.Errorf("err:%v", err)
				return
			}
		})
	})

}
