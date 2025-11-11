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
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/dungeonserver/internel/devent"
	"postapocgame/server/service/dungeonserver/internel/dungeonactor"
	"postapocgame/server/service/dungeonserver/internel/entity"
)

func init() {
	devent.Subscribe(1, func(ctx context.Context, event *event.Event) error {
		dungeonactor.RegisterFunc(protocol.RPC_EnterDungeon, func(msg *actor.Message) error {
			if msg.SessionId == "" {
				return customerr.NewCustomErr("not found session")
			}
			var roleInfo protocol.RoleInfo
			err := tool.JsonUnmarshal(msg.Data, &roleInfo)
			if err != nil {
				return customerr.Wrap(err)
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
				return customerr.Wrap(err)
			}
			return roleEntity.SendMessage(1, 3, bytes)
		})
		return nil
	})

}
