/**
 * @Author: zjj
 * @Date: 2025/11/10
 * @Desc:
**/

package fuben

import (
	"context"
	"postapocgame/server/internal"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/devent"
	"postapocgame/server/service/dungeonserver/internel/drpcprotocol"
	"postapocgame/server/service/dungeonserver/internel/dshare"
	"postapocgame/server/service/dungeonserver/internel/entity"
	"postapocgame/server/service/dungeonserver/internel/entitymgr"
	"postapocgame/server/service/dungeonserver/internel/gameserverlink"
)

func handleDoNetWorkMsg(actor.IActorMessage) {}
func handleDoRpcMsg(msg actor.IActorMessage) {
	req, err := dshare.Codec.DecodeRPCRequest(msg.GetData())
	if err != nil {
		return
	}
	getFunc := drpcprotocol.GetFunc(req.MsgId)
	if getFunc == nil {
		return
	}
	ctx := msg.GetContext()
	message := actor.NewBaseMessage(ctx, req.MsgId, req.Data)
	err = getFunc(message)
	if err != nil {
		return
	}
}

func handleG2DEnterDungeon(msg actor.IActorMessage) error {
	sessionId := msg.GetContext().Value(dshare.ContextKeySession).(string)
	if sessionId == "" {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "not found sessiopn")
	}
	var req protocol.G2DEnterDungeonReq
	err := internal.Unmarshal(msg.GetData(), &req)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	// 进入时绑定 session
	gameserverlink.GetMessageSender().RegisterSessionRoute(sessionId, req.PlatformId, req.SrvId)

	roleEntity := entity.NewRoleEntity(sessionId, req.SimpleData)
	err = entitymgr.GetEntityMgr().Register(roleEntity)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	resp := &protocol.S2CEnterSceneReq{
		EntityData: &protocol.EntitySt{
			Hdl:      roleEntity.GetHdl(),
			Id:       roleEntity.Id,
			Et:       roleEntity.GetEntityType(),
			PosX:     1,
			PosY:     1,
			SceneId:  1,
			FbId:     1,
			Level:    1,
			ShowName: req.SimpleData.RoleName,
		},
	}
	err = roleEntity.SendJsonMessage(uint16(protocol.S2CProtocol_S2CEnterScene), resp)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func init() {
	devent.Subscribe(devent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		dshare.RegisterHandler(uint16(dshare.DoNetWorkMsg), handleDoNetWorkMsg)
		dshare.RegisterHandler(uint16(dshare.DoRpcMsg), handleDoRpcMsg)
		drpcprotocol.Register(uint16(protocol.G2DRpcProtocol_G2DEnterDungeon), handleG2DEnterDungeon)
	})
}
