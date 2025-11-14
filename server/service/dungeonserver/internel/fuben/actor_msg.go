/**
 * @Author: zjj
 * @Date: 2025/11/10
 * @Desc:
**/

package fuben

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/clientprotocol"
	"postapocgame/server/service/dungeonserver/internel/devent"
	"postapocgame/server/service/dungeonserver/internel/drpcprotocol"
	"postapocgame/server/service/dungeonserver/internel/dshare"
	"postapocgame/server/service/dungeonserver/internel/entity"
	"postapocgame/server/service/dungeonserver/internel/entityhelper"
	"postapocgame/server/service/dungeonserver/internel/entitymgr"
	"postapocgame/server/service/dungeonserver/internel/gameserverlink"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"time"
)

func handleDoNetWorkMsg(msg actor.IActorMessage) {
	sessionId, _ := msg.GetContext().Value(dshare.ContextKeySession).(string)
	if sessionId == "" {
		log.Errorf("handleDoNetWorkMsg: missing session id")
		return
	}

	entityInstance, ok := entitymgr.GetEntityMgr().GetBySession(sessionId)
	if !ok {
		log.Warnf("handleDoNetWorkMsg: entity not found for session=%s", sessionId)
		return
	}

	cliMsg, err := dshare.Codec.DecodeClientMessage(msg.GetData())
	if err != nil {
		log.Errorf("handleDoNetWorkMsg decode failed: %v", err)
		return
	}

	handler := clientprotocol.GetFunc(cliMsg.MsgId)
	if handler == nil {
		log.Warnf("handleDoNetWorkMsg: no handler for proto=%d", cliMsg.MsgId)
		return
	}

	if err := handler(entityInstance, cliMsg); err != nil {
		log.Errorf("handleDoNetWorkMsg handler err: %v", err)
		_ = entityInstance.SendProtoMessage(uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  err.Error(),
		})
	}
}
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
	err := proto.Unmarshal(msg.GetData(), &req)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	// 进入时绑定 session
	gameserverlink.GetMessageSender().RegisterSessionRoute(sessionId, req.PlatformId, req.SrvId)

	// 创建角色实体，传入属性数据和技能列表
	roleEntity := entity.NewRoleEntity(sessionId, req.SimpleData, req.SyncAttrData, req.SkillMap)

	var fb iface.IFuBen
	var scene iface.IScene

	// 判断是进入默认副本还是限时副本
	if req.DungeonId == 0 {
		// 进入默认副本
		fb = GetDefaultFuBen()
		if fb == nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "default fuben not found")
		}
		scene = fb.GetScene(1)
		if scene == nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "default scene not found")
		}
	} else {
		// 进入限时副本
		var existingFb iface.IFuBen
		var exists bool
		if fbInstance, ok := getTimedFuBen(sessionId); ok {
			existingFb = fbInstance
			exists = true
		}
		if exists && existingFb != nil {
			// 使用已有副本
			fb = existingFb
		} else {
			// 创建新的限时副本
			// 获取副本配置
			configMgr := jsonconf.GetConfigManager()
			dungeonCfg, ok := configMgr.GetDungeonConfig(req.DungeonId)
			if !ok || dungeonCfg == nil {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "dungeon config not found")
			}

			// 限时副本默认1分钟
			maxDuration := 1 * time.Minute
			if dungeonCfg.Type != 2 {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "not a timed dungeon")
			}

			// 创建限时副本
			newFb, err := createTimedFuBen(sessionId, dungeonCfg.Name, maxDuration)
			if err != nil {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "failed to create timed fuben: %v", err)
			}
			fb = newFb

			// 设置难度
			fb.SetDifficulty(req.Difficulty)

			// 初始化场景（根据配置）
			sceneConfigs := make([]jsonconf.SceneConfig, 0)
			for _, sceneId := range dungeonCfg.SceneIds {
				sceneCfg, ok := configMgr.GetSceneConfig(sceneId)
				if ok && sceneCfg != nil {
					sceneConfigs = append(sceneConfigs, *sceneCfg)
				}
			}
			if len(sceneConfigs) == 0 {
				// 如果没有配置场景，使用默认场景
				sceneConfigs = []jsonconf.SceneConfig{
					{SceneId: 1, Name: "限时副本场景", Width: 1028, Height: 1028},
				}
			}
			fb.InitScenes(sceneConfigs)
		}

		// 玩家进入副本
		if err := fb.OnPlayerEnter(sessionId); err != nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "failed to enter fuben: %v", err)
		}

		// 获取第一个场景
		scenes := fb.GetAllScenes()
		if len(scenes) == 0 {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no scenes in timed fuben")
		}
		scene = scenes[0]
	}

	spawnX, spawnY := scene.GetRandomWalkablePos()
	roleEntity.SetPosition(spawnX, spawnY)

	if err := scene.AddEntity(roleEntity); err != nil {
		log.Errorf("add entity to scene failed: %v", err)
		return err
	}

	entitymgr.GetEntityMgr().BindSession(sessionId, roleEntity.GetHdl())
	if moveSys := roleEntity.GetMoveSys(); moveSys != nil {
		moveSys.ResetState()
	}

	resp := &protocol.S2CEnterSceneReq{
		EntityData: &protocol.EntitySt{
			Hdl:        roleEntity.GetHdl(),
			Id:         roleEntity.Id,
			Et:         roleEntity.GetEntityType(),
			PosX:       spawnX,
			PosY:       spawnY,
			SceneId:    scene.GetSceneId(),
			FbId:       fb.GetFbId(),
			Level:      roleEntity.GetLevel(),
			ShowName:   req.SimpleData.RoleName,
			Attrs:      entityhelper.BuildAttrMap(roleEntity),
			StateFlags: roleEntity.GetStateFlags(),
		},
	}
	if err := roleEntity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CEnterScene), resp); err != nil {
		log.Errorf("send enter scene failed: %v", err)
		return err
	}

	// 发送RPC通知GameServer进入副本成功
	notifyReq := &protocol.D2GEnterDungeonSuccessReq{
		SessionId: sessionId,
		RoleId:    req.SimpleData.RoleId,
	}
	notifyData, err := proto.Marshal(notifyReq)
	if err != nil {
		log.Errorf("marshal enter dungeon success notify failed: %v", err)
		// 不返回错误，继续执行
	} else {
		if err := gameserverlink.CallGameServer(context.Background(), sessionId, uint16(protocol.D2GRpcProtocol_D2GEnterDungeonSuccess), notifyData); err != nil {
			log.Errorf("send enter dungeon success notify failed: %v", err)
			// 不返回错误，继续执行
		} else {
			log.Infof("notified GameServer: role entered dungeon successfully, RoleId=%d", req.SimpleData.RoleId)
		}
	}

	return nil
}

// handleG2DSyncAttrs 处理属性同步请求
func handleG2DSyncAttrs(msg actor.IActorMessage) error {
	sessionId := msg.GetContext().Value(dshare.ContextKeySession).(string)
	if sessionId == "" {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "not found session")
	}

	var req protocol.G2DSyncAttrsReq
	err := proto.Unmarshal(msg.GetData(), &req)
	if err != nil {
		log.Errorf("unmarshal sync attrs request failed: %v", err)
		return err
	}

	// 获取角色实体
	entityInstance, ok := entitymgr.GetEntityMgr().GetBySession(sessionId)
	if !ok {
		log.Errorf("entity not found for session=%s", sessionId)
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "entity not found")
	}

	roleEntity, ok := entityInstance.(*entity.RoleEntity)
	if !ok {
		log.Errorf("entity is not RoleEntity, session=%s", sessionId)
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "entity type mismatch")
	}

	// 更新属性
	roleEntity.UpdateAttrs(req.SyncData)

	log.Debugf("synced attrs for role: RoleId=%d, Systems=%d", req.RoleId, len(req.SyncData.AttrData))
	return nil
}

// handleG2DUpdateHpMp 处理更新HP/MP请求
func handleG2DUpdateHpMp(msg actor.IActorMessage) error {
	sessionId := msg.GetContext().Value(dshare.ContextKeySession).(string)
	if sessionId == "" {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "not found session")
	}

	var req protocol.G2DUpdateHpMpReq
	err := proto.Unmarshal(msg.GetData(), &req)
	if err != nil {
		log.Errorf("unmarshal update hp/mp request failed: %v", err)
		return err
	}

	// 获取角色实体
	entityInstance, ok := entitymgr.GetEntityMgr().GetBySession(sessionId)
	if !ok {
		log.Errorf("entity not found for session=%s", sessionId)
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "entity not found")
	}

	roleEntity, ok := entityInstance.(*entity.RoleEntity)
	if !ok {
		log.Errorf("entity is not RoleEntity, session=%s", sessionId)
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "entity type mismatch")
	}

	// 更新HP/MP
	roleEntity.UpdateHpMp(req.HpDelta, req.MpDelta)

	log.Debugf("updated hp/mp for role: RoleId=%d, HPDelta=%d, MPDelta=%d", req.RoleId, req.HpDelta, req.MpDelta)
	return nil
}

// handleG2DUpdateSkill 处理更新技能请求（学习/升级）
func handleG2DUpdateSkill(msg actor.IActorMessage) error {
	sessionId := msg.GetContext().Value(dshare.ContextKeySession).(string)
	if sessionId == "" {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "not found session")
	}

	var req protocol.G2DUpdateSkillReq
	err := proto.Unmarshal(msg.GetData(), &req)
	if err != nil {
		log.Errorf("unmarshal update skill request failed: %v", err)
		return err
	}

	// 获取角色实体
	entityInstance, ok := entitymgr.GetEntityMgr().GetBySession(sessionId)
	if !ok {
		log.Errorf("entity not found for session=%s", sessionId)
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "entity not found")
	}

	roleEntity, ok := entityInstance.(*entity.RoleEntity)
	if !ok {
		log.Errorf("entity is not RoleEntity, session=%s", sessionId)
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "entity type mismatch")
	}

	// 更新技能
	err = roleEntity.UpdateSkill(req.SkillId, req.SkillLevel)
	if err != nil {
		log.Errorf("update skill failed: %v", err)
		return err
	}

	log.Debugf("updated skill for role: RoleId=%d, SkillId=%d, Level=%d", req.RoleId, req.SkillId, req.SkillLevel)
	return nil
}

func init() {
	devent.Subscribe(devent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		dshare.RegisterHandler(uint16(dshare.DoNetWorkMsg), handleDoNetWorkMsg)
		dshare.RegisterHandler(uint16(dshare.DoRpcMsg), handleDoRpcMsg)
		drpcprotocol.Register(uint16(protocol.G2DRpcProtocol_G2DEnterDungeon), handleG2DEnterDungeon)
		drpcprotocol.Register(uint16(protocol.G2DRpcProtocol_G2DSyncAttrs), handleG2DSyncAttrs)
		drpcprotocol.Register(uint16(protocol.G2DRpcProtocol_G2DUpdateHpMp), handleG2DUpdateHpMp)
		drpcprotocol.Register(uint16(protocol.G2DRpcProtocol_G2DUpdateSkill), handleG2DUpdateSkill)
	})
}
