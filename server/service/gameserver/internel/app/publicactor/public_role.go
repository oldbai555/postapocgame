package publicactor

import (
	"postapocgame/server/service/gameserver/internel/app/publicactor/offlinedata"
	"sync"
)

// PublicRole 公共角色（管理全局数据）
type PublicRole struct {
	// 在线状态管理：roleId -> sessionId
	onlineMap sync.Map // map[uint64]string

	// 排行榜数据：rankType -> RankData
	rankDataMap sync.Map // map[protocol.RankType]*protocol.RankData

	// 离线消息：roleId -> []ChatMessage
	offlineMessagesMap sync.Map // map[uint64][]*protocol.ChatMessage

	// 上次清理离线消息的时间（毫秒）
	lastCleanOfflineMessagesTime int64

	// OfflineData 管理器
	offlineDataMgr           *offlinedata.Manager
	lastOfflineDataFlushTime int64
}

// NewPublicRole 创建公共角色
func NewPublicRole() *PublicRole {
	return &PublicRole{
		onlineMap:      sync.Map{},
		rankDataMap:    sync.Map{},
		offlineDataMgr: offlinedata.NewManager(),
	}
}
