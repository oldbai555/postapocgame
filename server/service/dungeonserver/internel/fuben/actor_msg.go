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
		dshare.RegisterHandler(uint16(protocol.G2DRpcProtocol_G2DEnterDungeon), func(message actor.IActorMessage) {
			msg := message.(*base.SessionMessage)
			if msg.SessionId == "" {
				return
			}
			var roleInfo protocol.PlayerRoleData
			err := tool.JsonUnmarshal(msg.Data, &roleInfo)
			if err != nil {
				return
			}
			roleEntity := entity.NewRoleEntity(msg.SessionId, &roleInfo)
			resp := protocol.S2CEnterSceneReq{
				EntityData: &protocol.EntitySt{
					Hdl:      roleEntity.GetHdl(),
					Id:       roleEntity.Id,
					Et:       uint32(roleEntity.GetEntityType()),
					PosX:     1,
					PosY:     1,
					SceneId:  1,
					FbId:     1,
					Level:    1,
					ShowName: roleInfo.RoleName,
				},
			}
			bytes, err := tool.JsonMarshal(&resp)
			if err != nil {
				return
			}
			err = roleEntity.SendMessage(3, bytes)
			if err != nil {
				log.Errorf("err:%v", err)
				return
			}
		})
	})

}
