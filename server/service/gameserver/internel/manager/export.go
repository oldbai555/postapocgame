/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package manager

import (
	"postapocgame/server/service/gameserver/internel/iface"
)

// GetPlayerRole 获取玩家角色
func GetPlayerRole(playerRoleId uint64) iface.IPlayerRole {
	manager := GetPlayerRoleManager()
	playerRole, ok := manager.Get(playerRoleId)
	if !ok {
		return nil
	}
	return playerRole
}
