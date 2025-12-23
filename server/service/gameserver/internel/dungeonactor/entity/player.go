/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package entity

import (
	"context"
	"postapocgame/server/service/gameserver/internel/dungeonactor/iface"
	"postapocgame/server/service/gameserver/internel/gshare"
	"time"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
)

type Player struct {
	*BaseEntity
	sessionId string
	roleInfo  *protocol.PlayerSimpleData
	// 死亡相关
	dieTime time.Time // 死亡时间（用于延迟复活）
}

func NewPlayer(sessionId string, roleInfo *protocol.PlayerSimpleData, skillMap map[uint32]uint32) *Player {
	entity := &Player{
		BaseEntity: NewBaseEntity(roleInfo.RoleId, uint32(protocol.EntityType_EtPlayer)),
		sessionId:  sessionId,
		roleInfo:   roleInfo,
	}

	entity.GetAttrSys().SetAttrValue(attrdef.Level, int64(roleInfo.Level))

	// 初始化技能
	entity.initSkills(skillMap)
	entity.SetName(roleInfo.RoleName)

	return entity
}

func (r *Player) GetJobId() uint32 {
	if r.roleInfo == nil {
		return 0
	}
	return r.roleInfo.Job
}

func (r *Player) GetSessionId() string {
	return r.sessionId
}

func (r *Player) initSkills(skillMap map[uint32]uint32) {
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
func (r *Player) UpdateSkill(skillId, level uint32) error {

	// 如果技能已存在，先移除旧技能
	if r.GetFightSys().HasSkill(skillId) {
		// 重新学习技能（会更新等级）
		return r.GetFightSys().LearnSkill(skillId, level)
	}

	// 学习新技能
	return r.GetFightSys().LearnSkill(skillId, level)
}

func (r *Player) SendMessage(protoId uint16, data []byte) error {
	if r.sessionId == "" {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session id is empty")
	}

	req := &protocol.PAMSendToClientReq{
		MsgId: uint32(protoId),
		Data:  data,
	}
	payload, err := proto.Marshal(req)
	if err != nil {
		return customerr.Wrap(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, gshare.ContextKeySession, r.sessionId)
	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PlayerActorMsgId_PAMSendToClient), payload)
	return gshare.SendMessageAsync(r.sessionId, actorMsg)
}

func (r *Player) SendProtoMessage(protoId uint16, v proto.Message) error {
	bytes, err := proto.Marshal(v)
	if err != nil {
		return customerr.Wrap(err)
	}
	return r.SendMessage(protoId, bytes)
}

// OnDie 角色死亡处理（重写BaseEntity的方法）
func (r *Player) OnDie(killer iface.IEntity) {
	// 调用基类方法，设置死亡状态并广播
	r.BaseEntity.OnDie(killer)

	// 记录死亡时间（用于延迟复活，在RunOne中检查）
	r.dieTime = servertime.Now()
}
