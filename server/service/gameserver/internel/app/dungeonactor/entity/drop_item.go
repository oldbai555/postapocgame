package entity

import (
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entitymgr"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/iface"
	"time"
)

// DropItemEntity 掉落物实体
type DropItemEntity struct {
	*BaseEntity
	itemId      uint32        // 物品ID
	count       uint32        // 数量
	ownerHdl    uint64        // 归属者句柄（击杀怪物的玩家）
	ownerRoleId uint64        // 归属者角色ID（用于RPC调用）
	createTime  time.Time     // 创建时间
	lifetime    time.Duration // 存在时间（默认60秒）
}

// NewDropItemEntity 创建掉落物实体
func NewDropItemEntity(itemId, count uint32, x, y uint32, ownerHdl, ownerRoleId uint64, lifetimeSeconds uint32) *DropItemEntity {
	entity := &DropItemEntity{
		BaseEntity:  NewBaseEntity(0, uint32(protocol.EntityType_EtDropItem)),
		itemId:      itemId,
		count:       count,
		ownerHdl:    ownerHdl,
		ownerRoleId: ownerRoleId,
		createTime:  servertime.Now(),
		lifetime:    60 * time.Second, // 默认60秒
	}

	// 如果配置了存在时间，使用配置的时间
	if lifetimeSeconds > 0 {
		entity.lifetime = time.Duration(lifetimeSeconds) * time.Second
	}

	entity.SetPosition(x, y)

	// 掉落物不可被攻击、不可移动
	entity.SetInvincible(true)
	entity.SetCannotAttack(true)
	entity.SetCannotMove(true)

	return entity
}

// GetItemId 获取物品ID
func (d *DropItemEntity) GetItemId() uint32 {
	return d.itemId
}

// GetCount 获取数量
func (d *DropItemEntity) GetCount() uint32 {
	return d.count
}

// GetOwnerHdl 获取归属者句柄
func (d *DropItemEntity) GetOwnerHdl() uint64 {
	return d.ownerHdl
}

// GetOwnerRoleId 获取归属者角色ID
func (d *DropItemEntity) GetOwnerRoleId() uint64 {
	return d.ownerRoleId
}

// IsOwner 检查是否为归属者
func (d *DropItemEntity) IsOwner(entity iface.IEntity) bool {
	if entity == nil {
		return false
	}
	// 检查句柄是否匹配
	if entity.GetHdl() == d.ownerHdl {
		return true
	}
	// 如果是角色实体，检查角色ID是否匹配
	if roleEntity, ok := entity.(*RoleEntity); ok {
		return roleEntity.GetId() == d.ownerRoleId
	}
	return false
}

// RunOne 掉落物每帧更新（检查超时）
func (d *DropItemEntity) RunOne(now time.Time) {
	// 掉落物不需要调用BaseEntity的RunOne（不需要Buff、状态机、移动等）

	// 检查是否超时
	if now.Sub(d.createTime) >= d.lifetime {
		// 超时，从场景中移除
		d.removeFromScene()
	}
}

// removeFromScene 从场景中移除掉落物
func (d *DropItemEntity) removeFromScene() {
	hdl := d.GetHdl()
	entityMgr := entitymgr.GetEntityMgr()

	// 获取场景
	scene, ok := entityMgr.GetSceneByHandle(hdl)
	if ok && scene != nil {
		// 从场景中移除
		if err := scene.RemoveEntity(hdl); err != nil {
			log.Warnf("Failed to remove drop item from scene: %v", err)
		}
		// 解除场景绑定
		entityMgr.UnbindScene(hdl)
	}

	// 从实体管理器中注销
	entityMgr.Unregister(hdl)

	log.Infof("Drop item %d expired and removed", hdl)
}
