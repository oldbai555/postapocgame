// debuff_helper.go 提供 Buff/DOT 辅助方法，便于在 BuffSys 内复用。
package entitysystem

import "postapocgame/server/service/gameserver/internel/app/dungeonactor/iface"

// applyPeriodicDamage 对实体造成持续性伤害
func applyPeriodicDamage(entity iface.IEntity, damage int64) {
	if entity == nil || damage <= 0 {
		return
	}
	current := entity.GetHP()
	newHP := current - damage
	if newHP <= 0 {
		entity.SetHP(0)
		entity.OnDie(nil)
		return
	}
	entity.SetHP(newHP)
}
