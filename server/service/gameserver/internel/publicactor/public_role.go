package publicactor

import (
	"sync"

	"postapocgame/server/service/gameserver/internel/publicactor/offlinedata"
)

// PublicRole 公共角色（管理全局数据）
type PublicRole struct {
	// 在线状态管理：roleId -> sessionId
	onlineMap sync.Map // map[uint64]string

	// 排行榜数据：rankType -> RankData
	rankDataMap sync.Map // map[protocol.RankType]*protocol.RankData

	// 公会数据：guildId -> GuildData
	guildMap sync.Map // map[uint64]*protocol.GuildData

	// 拍卖行数据：auctionId -> AuctionItem
	auctionMap sync.Map // map[uint64]*protocol.AuctionItem

	// 离线消息：roleId -> []ChatMessage
	offlineMessagesMap sync.Map // map[uint64][]*protocol.ChatMessage

	// 公会申请：guildId -> []GuildApplication
	guildApplicationMap sync.Map // map[uint64][]*GuildApplication

	// 下一个公会ID（自增）
	nextGuildId uint64
	guildIdMu   sync.Mutex

	// 下一个拍卖ID（自增）
	nextAuctionId uint64
	auctionIdMu   sync.Mutex

	// 上次清理离线消息的时间（毫秒）
	lastCleanOfflineMessagesTime int64

	// OfflineData 管理器
	offlineDataMgr           *offlinedata.Manager
	lastOfflineDataFlushTime int64
}

// NewPublicRole 创建公共角色
func NewPublicRole() *PublicRole {
	return &PublicRole{
		onlineMap:           sync.Map{},
		rankDataMap:         sync.Map{},
		guildMap:            sync.Map{},
		auctionMap:          sync.Map{},
		nextGuildId:         1,
		nextAuctionId:       1,
		guildApplicationMap: sync.Map{},
		offlineDataMgr:      offlinedata.NewManager(),
	}
}
