package publicactor

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gshare"
)

var _ actor.IActorHandler = (*PublicHandler)(nil)

// NewPublicHandler 创建公共消息处理器
func NewPublicHandler() *PublicHandler {
	publicRole := NewPublicRole()
	handler := &PublicHandler{
		BaseActorHandler: actor.NewBaseActorHandler("public actor handler"),
		publicRole:       publicRole,
	}
	// 注意：RegisterMessageHandlers 需要在 PublicActor 创建后调用
	// 这里先不注册，等 PublicActor 创建后再注册
	// 加载数据
	handler.LoadData()
	return handler
}

// LoadData 从数据库加载公会和拍卖行数据
func (h *PublicHandler) LoadData() {
	// 加载公会数据
	guilds, err := database.GetAllGuilds()
	if err != nil {
		log.Warnf("Failed to load guilds from database: %v", err)
	} else {
		for _, guild := range guilds {
			h.publicRole.SetGuild(guild.GuildId, guild)
			// 更新nextGuildId
			if guild.GuildId >= h.publicRole.nextGuildId {
				h.publicRole.nextGuildId = guild.GuildId + 1
			}
		}
		log.Infof("Loaded %d guilds from database", len(guilds))
	}

	// 加载拍卖行数据
	auctionItems, err := database.GetAllAuctionItems()
	if err != nil {
		log.Warnf("Failed to load auction items from database: %v", err)
	} else {
		for _, item := range auctionItems {
			h.publicRole.SetAuctionItem(item.AuctionId, item)
			// 更新nextAuctionId
			if item.AuctionId >= h.publicRole.nextAuctionId {
				h.publicRole.nextAuctionId = item.AuctionId + 1
			}
		}
		log.Infof("Loaded %d auction items from database", len(auctionItems))
	}

	// 加载离线数据
	h.publicRole.LoadOfflineData(context.Background())
}

// PublicHandler 公共消息处理器
type PublicHandler struct {
	*actor.BaseActorHandler
	actorCtx   actor.IActorContext // 存储 Actor Context 引用
	publicRole *PublicRole         // 公共角色数据
}

// SetActorContext 设置 Actor Context 引用
func (h *PublicHandler) SetActorContext(ctx actor.IActorContext) {
	h.actorCtx = ctx
}

func (h *PublicHandler) Loop() {
	// Loop 方法只负责触发 RunOne，不注册处理器
	if h.actorCtx != nil {
		ctx := context.Background()
		h.actorCtx.ExecuteAsync(actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgRunOne), []byte{}))
	}
}

// OnStart Actor 启动时调用（只调用一次）
func (h *PublicHandler) OnStart() {
	// 注册 RunOne 消息处理器
	h.RegisterMessageHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgRunOne), func(msg actor.IActorMessage) {
		ctx := msg.GetContext()
		handleRunOneMsg(ctx, msg, h.publicRole)
	})

	// 触发第一次 RunOne
	if h.actorCtx != nil {
		ctx := context.Background()
		h.actorCtx.ExecuteAsync(actor.NewBaseMessage(ctx, gshare.DoRunOneMsg, []byte{}))
	}
}

// GetPublicRole 获取公共角色
func (h *PublicHandler) GetPublicRole() *PublicRole {
	return h.publicRole
}
