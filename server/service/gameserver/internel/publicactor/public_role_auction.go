package publicactor

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/manager"
)

// 拍卖行相关逻辑

// GetAuctionItem 获取拍卖物品
func (pr *PublicRole) GetAuctionItem(auctionId uint64) (*protocol.AuctionItem, bool) {
	value, ok := pr.auctionMap.Load(auctionId)
	if !ok {
		return nil, false
	}
	item, ok := value.(*protocol.AuctionItem)
	return item, ok
}

// SetAuctionItem 设置拍卖物品（并持久化到数据库）
func (pr *PublicRole) SetAuctionItem(auctionId uint64, item *protocol.AuctionItem) {
	pr.auctionMap.Store(auctionId, item)
	// 持久化到数据库
	if err := database.SaveAuctionItem(item); err != nil {
		log.Errorf("Failed to save auction item to database: %v", err)
	}
}

// DeleteAuctionItem 删除拍卖物品（并从数据库删除）
func (pr *PublicRole) DeleteAuctionItem(auctionId uint64) {
	pr.auctionMap.Delete(auctionId)
	// 从数据库删除
	if err := database.DeleteAuctionItem(auctionId); err != nil {
		log.Errorf("Failed to delete auction item from database: %v", err)
	}
}

// GetAllAuctionItems 获取所有拍卖物品（用于查询）
func (pr *PublicRole) GetAllAuctionItems() []*protocol.AuctionItem {
	var items []*protocol.AuctionItem
	pr.auctionMap.Range(func(key, value interface{}) bool {
		item, ok := value.(*protocol.AuctionItem)
		if ok {
			items = append(items, item)
		}
		return true
	})
	return items
}

// GetNextAuctionId 获取下一个拍卖ID
func (pr *PublicRole) GetNextAuctionId() uint64 {
	pr.auctionIdMu.Lock()
	defer pr.auctionIdMu.Unlock()
	id := pr.nextAuctionId
	pr.nextAuctionId++
	return id
}

// RunOne 每帧执行（处理过期拍卖、定期刷新等）
func (pr *PublicRole) RunOne(ctx context.Context) {
	// 处理过期的拍卖物品
	now := servertime.UnixMilli()
	pr.auctionMap.Range(func(key, value interface{}) bool {
		item, ok := value.(*protocol.AuctionItem)
		if ok && item.ExpireTime > 0 && item.ExpireTime < now {
			// 拍卖过期，需要返还物品给卖家（这里先删除，后续通过消息通知卖家）
			pr.auctionMap.Delete(key)
			log.Debugf("Auction item %d expired, removed", key)
		}
		return true
	})

	// 定期清理过期的离线消息（每小时执行一次）
	if pr.lastCleanOfflineMessagesTime == 0 || now-pr.lastCleanOfflineMessagesTime >= 3600000 { // 1小时 = 3600000毫秒
		if err := database.CleanExpiredOfflineMessages(); err != nil {
			log.Warnf("Failed to clean expired offline messages: %v", err)
		} else {
			log.Debugf("Cleaned expired offline messages")
		}
		pr.lastCleanOfflineMessagesTime = now
	}
}

// ===== 拍卖行相关 handler 注册（无闭包捕获 PublicRole） =====

// RegisterAuctionHandlers 注册拍卖行相关的消息处理器
func RegisterAuctionHandlers(facade gshare.IPublicActorFacade) {
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdAuctionPutOn), handleAuctionPutOnMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdAuctionBuy), handleAuctionBuyMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdAuctionQuery), handleAuctionQueryMsg)
}

// 拍卖行 handler 适配
var (
	handleAuctionPutOnMsg = withPublicRole(handleAuctionPutOn)
	handleAuctionBuyMsg   = withPublicRole(handleAuctionBuy)
	handleAuctionQueryMsg = withPublicRole(handleAuctionQuery)
)

// ===== 拍卖行业务 handler（从 message_handler.go 迁移）=====

// handleAuctionPutOn 处理拍卖上架
func handleAuctionPutOn(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	putOnMsg := &protocol.AuctionPutOnMsg{}
	if err := proto.Unmarshal(data, putOnMsg); err != nil {
		log.Errorf("Failed to unmarshal AuctionPutOnMsg: %v", err)
		return
	}

	// 生成拍卖ID
	auctionId := publicRole.GetNextAuctionId()
	now := servertime.UnixMilli()
	expireTime := now + putOnMsg.Duration*1000 // duration是秒，转换为毫秒

	// 创建拍卖物品
	auctionItem := &protocol.AuctionItem{
		AuctionId:  auctionId,
		ItemId:     putOnMsg.ItemId,
		Count:      putOnMsg.Count,
		Price:      putOnMsg.Price,
		SellerId:   putOnMsg.SellerId,
		ExpireTime: expireTime,
		CreateTime: now,
	}

	// 存储拍卖物品
	publicRole.SetAuctionItem(auctionId, auctionItem)

	// 发送响应给卖家
	_, ok := publicRole.GetSessionId(putOnMsg.SellerId)
	if ok {
		// TODO: 添加S2CAuctionPutOnResult协议定义后发送响应
		log.Debugf("Auction put on success: auctionId=%d", auctionId)
	}

	log.Debugf("handleAuctionPutOn: item %d put on auction %d by seller %d", putOnMsg.ItemId, auctionId, putOnMsg.SellerId)
}

// handleAuctionBuy 处理拍卖购买
func handleAuctionBuy(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	buyMsg := &protocol.AuctionBuyMsg{}
	if err := proto.Unmarshal(data, buyMsg); err != nil {
		log.Errorf("Failed to unmarshal AuctionBuyMsg: %v", err)
		return
	}

	// 查找拍卖物品
	auctionItem, ok := publicRole.GetAuctionItem(buyMsg.AuctionId)
	if !ok {
		log.Warnf("handleAuctionBuy: auction %d not found", buyMsg.AuctionId)
		// 发送失败响应
		buyerSessionId, ok := publicRole.GetSessionId(buyMsg.BuyerId)
		if ok {
			respMsg := &protocol.S2CAuctionBuyResultReq{
				Success: false,
				Message: "拍卖物品不存在",
			}
			respData, err := proto.Marshal(respMsg)
			if err == nil {
				gatewaylink.SendToSession(buyerSessionId, uint16(protocol.S2CProtocol_S2CAuctionBuyResult), respData)
			}
		}
		return
	}

	// 检查是否过期
	now := servertime.UnixMilli()
	if auctionItem.ExpireTime > 0 && auctionItem.ExpireTime < now {
		log.Warnf("handleAuctionBuy: auction %d expired", buyMsg.AuctionId)
		publicRole.DeleteAuctionItem(buyMsg.AuctionId)
		// 发送失败响应
		buyerSessionId, ok := publicRole.GetSessionId(buyMsg.BuyerId)
		if ok {
			respMsg := &protocol.S2CAuctionBuyResultReq{
				Success: false,
				Message: "拍卖物品已过期",
			}
			respData, err := proto.Marshal(respMsg)
			if err == nil {
				gatewaylink.SendToSession(buyerSessionId, uint16(protocol.S2CProtocol_S2CAuctionBuyResult), respData)
			}
		}
		return
	}

	// 检查是否是自己上架的
	if auctionItem.SellerId == buyMsg.BuyerId {
		log.Warnf("handleAuctionBuy: buyer %d cannot buy own item", buyMsg.BuyerId)
		buyerSessionId, ok := publicRole.GetSessionId(buyMsg.BuyerId)
		if ok {
			respMsg := &protocol.S2CAuctionBuyResultReq{
				Success: false,
				Message: "不能购买自己上架的物品",
			}
			respData, err := proto.Marshal(respMsg)
			if err == nil {
				gatewaylink.SendToSession(buyerSessionId, uint16(protocol.S2CProtocol_S2CAuctionBuyResult), respData)
			}
		}
		return
	}

	// 通过manager获取买家和卖家的PlayerRole
	buyerRole := manager.GetPlayerRole(buyMsg.BuyerId)
	if buyerRole == nil {
		log.Warnf("handleAuctionBuy: buyer %d not found", buyMsg.BuyerId)
		buyerSessionId, ok := publicRole.GetSessionId(buyMsg.BuyerId)
		if ok {
			respMsg := &protocol.S2CAuctionBuyResultReq{
				Success: false,
				Message: "买家不在线",
			}
			respData, err := proto.Marshal(respMsg)
			if err == nil {
				gatewaylink.SendToSession(buyerSessionId, uint16(protocol.S2CProtocol_S2CAuctionBuyResult), respData)
			}
		}
		return
	}

	// 扣除买家货币（默认使用金币，货币ID=2）
	buyerCtx := buyerRole.WithContext(nil)
	moneyId := uint32(protocol.MoneyType_MoneyTypeGoldCoin) // 默认金币
	consumeItems := []*jsonconf.ItemAmount{
		{
			ItemId:   moneyId,
			Count:    auctionItem.Price,
			ItemType: uint32(protocol.ItemType_ItemTypeMoney),
		},
	}

	// 检查货币是否足够
	if err := buyerRole.CheckConsume(buyerCtx, consumeItems); err != nil {
		log.Warnf("handleAuctionBuy: buyer %d money not enough: %v", buyMsg.BuyerId, err)
		buyerSessionId, ok := publicRole.GetSessionId(buyMsg.BuyerId)
		if ok {
			respMsg := &protocol.S2CAuctionBuyResultReq{
				Success: false,
				Message: "货币不足",
			}
			respData, err := proto.Marshal(respMsg)
			if err == nil {
				gatewaylink.SendToSession(buyerSessionId, uint16(protocol.S2CProtocol_S2CAuctionBuyResult), respData)
			}
		}
		return
	}

	// 执行扣除货币
	if err := buyerRole.ApplyConsume(buyerCtx, consumeItems); err != nil {
		log.Errorf("handleAuctionBuy: failed to deduct money from buyer %d: %v", buyMsg.BuyerId, err)
		buyerSessionId, ok := publicRole.GetSessionId(buyMsg.BuyerId)
		if ok {
			respMsg := &protocol.S2CAuctionBuyResultReq{
				Success: false,
				Message: "扣除货币失败",
			}
			respData, err := proto.Marshal(respMsg)
			if err == nil {
				gatewaylink.SendToSession(buyerSessionId, uint16(protocol.S2CProtocol_S2CAuctionBuyResult), respData)
			}
		}
		return
	}

	// 给买家添加物品
	rewardItems := []*jsonconf.ItemAmount{
		{
			ItemId:   auctionItem.ItemId,
			Count:    int64(auctionItem.Count),
			ItemType: 0, // 普通物品
		},
	}
	if err := buyerRole.GrantRewards(buyerCtx, rewardItems); err != nil {
		log.Errorf("handleAuctionBuy: failed to add item to buyer %d: %v", buyMsg.BuyerId, err)
		buyerSessionId, ok := publicRole.GetSessionId(buyMsg.BuyerId)
		if ok {
			respMsg := &protocol.S2CAuctionBuyResultReq{
				Success: false,
				Message: "添加物品失败",
			}
			respData, err := proto.Marshal(respMsg)
			if err == nil {
				gatewaylink.SendToSession(buyerSessionId, uint16(protocol.S2CProtocol_S2CAuctionBuyResult), respData)
			}
		}
		return
	}

	// 给卖家添加货币
	sellerRole := manager.GetPlayerRole(auctionItem.SellerId)
	if sellerRole != nil {
		sellerCtx := sellerRole.WithContext(nil)
		sellerRewards := []*jsonconf.ItemAmount{
			{
				ItemId:   moneyId,
				Count:    auctionItem.Price,
				ItemType: uint32(protocol.ItemType_ItemTypeMoney),
			},
		}
		if err := sellerRole.GrantRewards(sellerCtx, sellerRewards); err != nil {
			log.Errorf("handleAuctionBuy: failed to add money to seller %d: %v", auctionItem.SellerId, err)
		}
	}

	// 交易审计
	transactionStatus := uint32(1) // 1=成功
	transactionReason := "交易成功"
	if auctionItem.Price < 0 || auctionItem.Price > 1000000000 {
		transactionStatus = uint32(3) // 3=可疑
		transactionReason = "价格异常"
	}
	if err := database.SaveTransactionAudit(
		uint32(1),
		buyMsg.BuyerId,
		auctionItem.SellerId,
		auctionItem.ItemId,
		auctionItem.Count,
		auctionItem.Price,
		transactionStatus,
		transactionReason,
	); err != nil {
		log.Warnf("handleAuctionBuy: failed to save transaction audit: %v", err)
	}

	// 删除拍卖物品
	publicRole.DeleteAuctionItem(buyMsg.AuctionId)

	// 发送成功响应给买家
	buyerSessionId, ok := publicRole.GetSessionId(buyMsg.BuyerId)
	if ok {
		respMsg := &protocol.S2CAuctionBuyResultReq{
			Success: true,
			Message: "购买成功",
			Item:    auctionItem,
		}
		respData, err := proto.Marshal(respMsg)
		if err == nil {
			gatewaylink.SendToSession(buyerSessionId, uint16(protocol.S2CProtocol_S2CAuctionBuyResult), respData)
		}
	}

	log.Debugf("handleAuctionBuy: buyer %d bought auction %d", buyMsg.BuyerId, buyMsg.AuctionId)
}

// handleAuctionQuery 处理拍卖查询
func handleAuctionQuery(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	queryMsg := &protocol.AuctionQueryMsg{}
	if err := proto.Unmarshal(data, queryMsg); err != nil {
		log.Errorf("Failed to unmarshal AuctionQueryMsg: %v", err)
		return
	}

	// 获取所有拍卖物品
	allItems := publicRole.GetAllAuctionItems()
	now := servertime.UnixMilli()

	// 过滤过期物品和按条件筛选
	var filteredItems []*protocol.AuctionItem
	for _, item := range allItems {
		if item.ExpireTime > 0 && item.ExpireTime < now {
			continue
		}
		if queryMsg.ItemId > 0 && item.ItemId != queryMsg.ItemId {
			continue
		}
		filteredItems = append(filteredItems, item)
	}

	// 分页处理
	page := int(queryMsg.Page)
	pageSize := int(queryMsg.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	totalCount := len(filteredItems)
	totalPage := (totalCount + pageSize - 1) / pageSize
	if totalPage == 0 {
		totalPage = 1
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	var resultItems []*protocol.AuctionItem
	if start < end {
		resultItems = filteredItems[start:end]
	}

	// 构建响应
	respMsg := &protocol.S2CAuctionQueryResultReq{
		Items:      resultItems,
		TotalCount: int32(totalCount),
		TotalPage:  int32(totalPage),
	}
	respData, err := proto.Marshal(respMsg)
	if err != nil {
		log.Errorf("Failed to marshal S2CAuctionQueryResultReq: %v", err)
		return
	}

	// 发送给请求者
	err = gatewaylink.SendToSession(queryMsg.RequesterSessionId, uint16(protocol.S2CProtocol_S2CAuctionQueryResult), respData)
	if err != nil {
		log.Warnf("Failed to send auction query result to session %s: %v", queryMsg.RequesterSessionId, err)
	}

	log.Debugf("handleAuctionQuery: sent %d items (page %d/%d) to session %s", len(resultItems), page, totalPage, queryMsg.RequesterSessionId)
}
