package entitysystem

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/playeractor/clientprotocol"
)

// AuctionSys 拍卖行系统
type AuctionSys struct {
	*BaseSystem
	data *protocol.SiAuctionData
}

// NewAuctionSys 创建拍卖行系统
func NewAuctionSys() iface.ISystem {
	return &AuctionSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysAuction)),
	}
}

func (s *AuctionSys) OnInit(ctx context.Context) {
	role, err := GetIPlayerRoleByContext(ctx)
	if err != nil || role == nil {
		return
	}
	bd := role.GetBinaryData()
	if bd.AuctionData == nil {
		bd.AuctionData = &protocol.SiAuctionData{
			AuctionIdList: make([]uint64, 0),
		}
	}
	s.data = bd.AuctionData
}

// GetAuctionIdList 获取拍卖ID列表
func (s *AuctionSys) GetAuctionIdList() []uint64 {
	if s.data == nil {
		return nil
	}
	return s.data.AuctionIdList
}

// AddAuctionId 添加拍卖ID
func (s *AuctionSys) AddAuctionId(auctionId uint64) bool {
	if s.data == nil {
		return false
	}
	// 检查是否已经存在
	for _, id := range s.data.AuctionIdList {
		if id == auctionId {
			return false
		}
	}
	s.data.AuctionIdList = append(s.data.AuctionIdList, auctionId)
	return true
}

// RemoveAuctionId 移除拍卖ID
func (s *AuctionSys) RemoveAuctionId(auctionId uint64) bool {
	if s.data == nil {
		return false
	}
	for i, id := range s.data.AuctionIdList {
		if id == auctionId {
			s.data.AuctionIdList = append(s.data.AuctionIdList[:i], s.data.AuctionIdList[i+1:]...)
			return true
		}
	}
	return false
}

// handleAuctionPutOn 处理拍卖上架
func handleAuctionPutOn(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleAuctionPutOn: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SAuctionPutOnReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleAuctionPutOn: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()
	putOnMsg := &protocol.AuctionPutOnMsg{
		SellerId: roleId,
		ItemId:   req.ItemId,
		Count:    req.Count,
		Price:    req.Price,
		Duration: req.Duration,
	}

	msgData, err := proto.Marshal(putOnMsg)
	if err != nil {
		log.Errorf("handleAuctionPutOn: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAuctionPutOn), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleAuctionPutOn: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// handleAuctionBuy 处理拍卖购买
func handleAuctionBuy(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleAuctionBuy: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SAuctionBuyReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleAuctionBuy: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()
	buyMsg := &protocol.AuctionBuyMsg{
		BuyerId:   roleId,
		AuctionId: req.AuctionId,
	}

	msgData, err := proto.Marshal(buyMsg)
	if err != nil {
		log.Errorf("handleAuctionBuy: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAuctionBuy), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleAuctionBuy: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// handleAuctionQuery 处理拍卖查询
func handleAuctionQuery(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	_, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleAuctionQuery: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SAuctionQueryReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleAuctionQuery: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	queryMsg := &protocol.AuctionQueryMsg{
		ItemId:             req.ItemId,
		Page:               req.Page,
		PageSize:           req.PageSize,
		RequesterSessionId: sessionId,
	}

	msgData, err := proto.Marshal(queryMsg)
	if err != nil {
		log.Errorf("handleAuctionQuery: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAuctionQuery), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleAuctionQuery: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysAuction), NewAuctionSys)
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SAuctionPutOn), handleAuctionPutOn)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SAuctionBuy), handleAuctionBuy)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SAuctionQuery), handleAuctionQuery)
	})
}
