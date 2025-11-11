/**
 * @Author: zjj
 * @Date: 2025/11/6
 * @Desc:
**/

package argsdef

import "postapocgame/server/internal/network"

// GameServerInfo GameServer连接信息
type GameServerInfo struct {
	PlatformId uint32
	ZoneId     uint32
	Conn       network.IConnection
}

// GameServerKey GameServer唯一标识
type GameServerKey struct {
	PlatformId uint32
	ZoneId     uint32
}
