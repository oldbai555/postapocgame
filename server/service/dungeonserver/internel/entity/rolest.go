/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package entity

import (
	"context"
	"math/rand"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/network"
	"postapocgame/server/service/dungeonserver/internel/clientprotocol"
	"postapocgame/server/service/dungeonserver/internel/devent"
	"time"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/entitymgr"
	"postapocgame/server/service/dungeonserver/internel/gameserverlink"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

type attrSyncApplier interface {
	ApplySyncData(*protocol.SyncAttrData)
}

// RoleEntity 角色实体
type RoleEntity struct {
	*BaseEntity
	sessionId string
	roleInfo  *protocol.PlayerSimpleData
	// 死亡相关
	dieTime time.Time // 死亡时间（用于延迟复活）
}

// NewRoleEntity 创建角色实体
func NewRoleEntity(sessionId string, roleInfo *protocol.PlayerSimpleData, syncAttrData *protocol.SyncAttrData, skillMap map[uint32]uint32) *RoleEntity {
	entity := &RoleEntity{
		BaseEntity: NewBaseEntity(roleInfo.RoleId, uint32(protocol.EntityType_EtRole)),
		sessionId:  sessionId,
		roleInfo:   roleInfo,
	}

	entity.GetAttrSys().SetAttrValue(attrdef.AttrLevel, attrdef.AttrValue(roleInfo.Level))
	entity.initAttrSys(syncAttrData)

	// 初始化技能
	entity.initSkills(skillMap)
	entity.SetName(roleInfo.RoleName)

	// 标记属性系统初始化完成（允许广播属性）
	entity.GetAttrSys().SetInitFinish()

	return entity
}

func (r *RoleEntity) initAttrSys(syncAttrData *protocol.SyncAttrData) {
	attrSys := r.GetAttrSys()
	// 如果有传入的属性数据，使用传入的属性数据初始化
	if syncAttrData != nil {
		if applier, ok := attrSys.(attrSyncApplier); ok {
			applier.ApplySyncData(syncAttrData)
		}
	} else {
		// 如果没有传入属性数据，使用等级配置表初始化（兼容旧逻辑）
		level := attrSys.GetAttrValue(attrdef.AttrLevel)
		levelAttrs := jsonconf.GetConfigManager().GetLevelAttrs(uint32(level))
		for attrType, attrValue := range levelAttrs {
			attrSys.SetAttrValue(attrType, attrdef.AttrValue(attrValue))
		}
	}

	// 如果配置中没有设置HP/MP的当前值，则设置为最大值
	if attrSys.GetAttrValue(attrdef.AttrHP) == 0 {
		maxHP := attrSys.GetAttrValue(attrdef.AttrMaxHP)
		if maxHP > 0 {
			attrSys.SetAttrValue(attrdef.AttrHP, maxHP)
		}
	}
	if attrSys.GetAttrValue(attrdef.AttrMP) == 0 {
		maxMP := attrSys.GetAttrValue(attrdef.AttrMaxMP)
		if maxMP > 0 {
			attrSys.SetAttrValue(attrdef.AttrMP, maxMP)
		}
	}
}

// UpdateAttrs 更新属性（对比差异，增加和减少属性）
func (r *RoleEntity) UpdateAttrs(newSyncData *protocol.SyncAttrData) {
	if newSyncData == nil {
		return
	}

	if applier, ok := r.GetAttrSys().(attrSyncApplier); ok {
		applier.ApplySyncData(newSyncData)
	}
}

func (r *RoleEntity) GetJobId() uint32 {
	if r.roleInfo == nil {
		return 0
	}
	return r.roleInfo.Job
}

func (r *RoleEntity) GetSessionId() string {
	return r.sessionId
}

func (r *RoleEntity) initSkills(skillMap map[uint32]uint32) {
	if skillMap == nil {
		return
	}
	// 使用传入的技能列表初始化
	for skillId, level := range skillMap {
		err := r.GetFightSys().LearnSkill(skillId, level)
		if err != nil {
			log.Errorf("init skill %d %d failed: %v", skillId, level, err)
		}
	}
}

// UpdateSkill 更新技能（学习或升级）
func (r *RoleEntity) UpdateSkill(skillId, level uint32) error {
	if r.GetFightSys() == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "fight system not found")
	}

	// 如果技能已存在，先移除旧技能
	if r.GetFightSys().HasSkill(skillId) {
		// 重新学习技能（会更新等级）
		return r.GetFightSys().LearnSkill(skillId, level)
	}

	// 学习新技能
	return r.GetFightSys().LearnSkill(skillId, level)
}

func (r *RoleEntity) SendMessage(protoId uint16, data []byte) error {
	return gameserverlink.SendToClient(r.sessionId, protoId, data)
}

func (r *RoleEntity) SendProtoMessage(protoId uint16, v proto.Message) error {
	bytes, err := proto.Marshal(v)
	if err != nil {
		return customerr.Wrap(err)
	}
	return r.SendMessage(protoId, bytes)
}

// UpdateHpMp 更新HP/MP（用于物品使用等场景）
func (r *RoleEntity) UpdateHpMp(hpDelta int64, mpDelta int64) {
	attrSys := r.GetAttrSys()

	// 更新HP
	if hpDelta != 0 {
		currentHP := attrSys.GetAttrValue(attrdef.AttrHP)
		maxHP := attrSys.GetAttrValue(attrdef.AttrMaxHP)
		newHP := currentHP + attrdef.AttrValue(hpDelta)
		// 限制在0到最大值之间
		if newHP < 0 {
			newHP = 0
		} else if newHP > maxHP {
			newHP = maxHP
		}
		attrSys.SetAttrValue(attrdef.AttrHP, newHP)
	}

	// 更新MP
	if mpDelta != 0 {
		currentMP := attrSys.GetAttrValue(attrdef.AttrMP)
		maxMP := attrSys.GetAttrValue(attrdef.AttrMaxMP)
		newMP := currentMP + attrdef.AttrValue(mpDelta)
		// 限制在0到最大值之间
		if newMP < 0 {
			newMP = 0
		} else if newMP > maxMP {
			newMP = maxMP
		}
		attrSys.SetAttrValue(attrdef.AttrMP, newMP)
	}
}

// RunOne 每帧更新（重写BaseEntity的方法）
func (r *RoleEntity) RunOne(now time.Time) {
	// 调用基类方法
	r.BaseEntity.RunOne(now)

	// 检查是否需要复活（死亡3秒后自动复活）
	if !r.dieTime.IsZero() && r.IsDead() {
		if now.Sub(r.dieTime) >= 3*time.Second {
			r.Revive()
		}
	}
}

// OnDie 角色死亡处理（重写BaseEntity的方法）
func (r *RoleEntity) OnDie(killer iface.IEntity) {
	// 调用基类方法，设置死亡状态并广播
	r.BaseEntity.OnDie(killer)

	// 记录死亡时间（用于延迟复活，在RunOne中检查）
	r.dieTime = servertime.Now()
}

// Revive 复活角色到新手村
func (r *RoleEntity) Revive() error {
	// 获取复活场景（默认新手村场景ID=1）
	newbieScene := getReviveScene()
	if newbieScene == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "newbie scene not found")
	}

	// 获取当前场景
	if currentScene, ok := entitymgr.GetEntityMgr().GetSceneByHandle(r.GetHdl()); ok && currentScene != nil {
		// 从当前场景移除
		currentScene.RemoveEntity(r.GetHdl())
	}

	// 获取新手村出生点
	configMgr := jsonconf.GetConfigManager()
	sceneConfig, _ := configMgr.GetSceneConfig(1)
	var x, y uint32 = 100, 100 // 默认位置
	if sceneConfig != nil && sceneConfig.BornArea != nil {
		bornArea := sceneConfig.BornArea
		if bornArea.X2 > bornArea.X1 && bornArea.Y2 > bornArea.Y1 {
			// 从出生点范围随机选择
			x = bornArea.X1 + uint32(rand.Intn(int(bornArea.X2-bornArea.X1)))
			y = bornArea.Y1 + uint32(rand.Intn(int(bornArea.Y2-bornArea.Y1)))
		}
	}

	// 设置位置
	r.SetPosition(x, y)

	// 添加到新手村场景
	if err := newbieScene.AddEntity(r); err != nil {
		return customerr.Wrap(err)
	}

	// 恢复满血满蓝
	attrSys := r.GetAttrSys()
	maxHP := attrSys.GetAttrValue(attrdef.AttrMaxHP)
	maxMP := attrSys.GetAttrValue(attrdef.AttrMaxMP)
	attrSys.SetAttrValue(attrdef.AttrHP, maxHP)
	attrSys.SetAttrValue(attrdef.AttrMP, maxMP)

	// 清除死亡状态（不需要加锁，因为Actor模型保证单线程）
	r.stateFlags &^= (uint64(1) << uint(protocol.EntityStateFlag_EntityStateFlagDead))
	r.dieTime = time.Time{} // 清除死亡时间

	// 发送复活结果
	resp := &protocol.S2CReviveResultReq{
		Success: true,
		Message: "复活成功",
		SceneId: 1,
		PosX:    x,
		PosY:    y,
	}
	r.SendProtoMessage(uint16(protocol.S2CProtocol_S2CReviveResult), resp)

	log.Infof("Role %d revived at scene 1, pos=(%d,%d)", r.GetId(), x, y)
	return nil
}

func handleRevive(role iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SReviveReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// 检查实体是否为角色实体
	roleEntity, ok := role.(*RoleEntity)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "entity is not a role entity")
	}

	// 检查角色是否死亡
	if !roleEntity.IsDead() {
		resp := &protocol.S2CReviveResultReq{
			Success: false,
			Message: "角色未死亡，无需复活",
		}
		return roleEntity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CReviveResult), resp)
	}

	// 执行复活
	err := roleEntity.Revive()
	if err != nil {
		resp := &protocol.S2CReviveResultReq{
			Success: false,
			Message: err.Error(),
		}
		return roleEntity.SendProtoMessage(uint16(protocol.S2CProtocol_S2CReviveResult), resp)
	}

	// Revive方法内部已经发送了响应，这里不需要再发送
	return nil
}

func init() {
	devent.Subscribe(devent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SRevive), handleRevive)
	})
}
