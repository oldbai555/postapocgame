/**
 * @Author: zjj
 * @Date: 2025/11/10
 * @Desc:
**/

package fuben

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entity"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entitymgr"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/iface"
	gshare "postapocgame/server/service/gameserver/internel/core/gshare"
)

// HandleG2DEnterDungeon 处理进入副本请求（GameServer → DungeonActor 内部消息）
func HandleG2DEnterDungeon(msg actor.IActorMessage) error {
	sessionId, _ := msg.GetContext().Value("session").(string)
	if sessionId == "" {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "not found sessiopn")
	}
	var req protocol.G2DEnterDungeonReq
	err := proto.Unmarshal(msg.GetData(), &req)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

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

	spawnX, spawnY := scene.GetSpawnPos()
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
		EntityData: roleEntity.BuildProtoEntitySt(),
	}
	if err := roleEntity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CEnterScene), resp); err != nil {
		log.Errorf("send enter scene failed: %v", err)
		return err
	}

	// 发送进入副本成功消息给 PlayerActor
	sendEnterDungeonSuccessToPlayerActor(sessionId, req.SimpleData.RoleId)

	return nil
}

// HandleG2DSyncAttrs 处理属性同步请求（GameServer → DungeonActor 内部消息）
func HandleG2DSyncAttrs(msg actor.IActorMessage) error {
	sessionId, _ := msg.GetContext().Value("session").(string)
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

// HandleG2DUpdateHpMp 处理更新HP/MP请求（GameServer → DungeonActor 内部消息）
func HandleG2DUpdateHpMp(msg actor.IActorMessage) error {
	sessionId, _ := msg.GetContext().Value("session").(string)
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

// HandleG2DUpdateSkill 处理更新技能请求（学习/升级，GameServer → DungeonActor 内部消息）
func HandleG2DUpdateSkill(msg actor.IActorMessage) error {
	sessionId, _ := msg.GetContext().Value("session").(string)
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

// sendEnterDungeonSuccessToPlayerActor 发送进入副本成功消息给 PlayerActor
func sendEnterDungeonSuccessToPlayerActor(sessionId string, roleId uint64) {
	req := &protocol.D2GEnterDungeonSuccessReq{
		SessionId: sessionId,
		RoleId:    roleId,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		log.Errorf("[actor_msg] marshal D2GEnterDungeonSuccessReq failed: %v", err)
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, gshare.ContextKeySession, sessionId)
	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdEnterDungeonSuccess), data)
	if err := gshare.SendMessageAsync(sessionId, actorMsg); err != nil {
		log.Errorf("[actor_msg] send EnterDungeonSuccess message to PlayerActor failed: sessionId=%s err=%v", sessionId, err)
	}
}
