package entitysystem

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
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

// ShopSys 个人商城系统
type ShopSys struct {
	*BaseSystem
	purchaseCounters map[uint32]uint32
}

func NewShopSys() *ShopSys {
	return &ShopSys{
		BaseSystem:       NewBaseSystem(uint32(protocol.SystemId_SysShop)),
		purchaseCounters: make(map[uint32]uint32),
	}
}

func GetShopSys(ctx context.Context) *ShopSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get ShopSys player role failed: %v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysShop))
	if system == nil {
		return nil
	}
	shopSys, ok := system.(*ShopSys)
	if !ok || !shopSys.IsOpened() {
		return nil
	}
	return shopSys
}

// Buy 购买商城商品
func (ss *ShopSys) Buy(ctx context.Context, itemID, count uint32) error {
	if count == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "count invalid")
	}
	cfg, ok := jsonconf.GetConfigManager().GetItemConfig(itemID)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "item config not found")
	}

	consumes, err := ss.buildConsumeList(cfg, count)
	if err != nil {
		return err
	}
	rewards := ss.buildRewardList(cfg, count)

	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	if err := playerRole.ApplyConsume(ctx, consumes); err != nil {
		return err
	}
	if err := playerRole.GrantRewards(ctx, rewards); err != nil {
		return err
	}

	ss.recordPurchase(itemID, count)
	return nil
}

func (ss *ShopSys) buildConsumeList(cfg *jsonconf.ItemConfig, count uint32) ([]*jsonconf.ItemAmount, error) {
	if cfg == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config nil")
	}
	manager := jsonconf.GetConfigManager()

	// 从商城配置读取消耗
	shopCfg, ok := manager.GetShopConfig(cfg.ItemId)
	if ok && shopCfg != nil {
		if shopCfg.ConsumeId > 0 {
			if consumeCfg, ok := manager.GetConsumeConfig(shopCfg.ConsumeId); ok && consumeCfg != nil {
				return scaleItemAmounts(consumeCfg.Items, count), nil
			}
		}
	}

	// 如果没有商城配置，返回错误（必须通过商城配置购买）
	return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "item not in shop")
}

func (ss *ShopSys) buildRewardList(cfg *jsonconf.ItemConfig, count uint32) []*jsonconf.ItemAmount {
	if cfg == nil {
		return nil
	}
	manager := jsonconf.GetConfigManager()

	// 从商城配置读取奖励
	shopCfg, ok := manager.GetShopConfig(cfg.ItemId)
	if ok && shopCfg != nil {
		if shopCfg.RewardId > 0 {
			if rewardCfg, ok := manager.GetRewardConfig(shopCfg.RewardId); ok && rewardCfg != nil {
				return scaleItemAmounts(rewardCfg.Items, count)
			}
		}
	}

	// 如果没有商城配置，返回默认奖励（物品本身）
	return []*jsonconf.ItemAmount{
		{
			ItemType: cfg.Type,
			ItemId:   cfg.ItemId,
			Count:    int64(count),
			Bind:     1,
		},
	}
}

func (ss *ShopSys) recordPurchase(itemID, count uint32) {
	ss.purchaseCounters[itemID] += count
}

func scaleItemAmounts(items []*jsonconf.ItemAmount, times uint32) []*jsonconf.ItemAmount {
	if times == 0 || len(items) == 0 {
		return nil
	}
	result := make([]*jsonconf.ItemAmount, 0, len(items))
	for _, item := range items {
		if item == nil || item.Count <= 0 {
			continue
		}
		cp := item.Clone()
		if cp == nil {
			continue
		}
		cp.Count = cp.Count * int64(times)
		result = append(result, cp)
	}
	return result
}

func handleShopBuy(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	var req protocol.C2SShopBuyReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}
	resp := &protocol.S2CShopBuyResultReq{
		ItemId: req.ItemId,
		Count:  req.Count,
	}
	sendResp := func() error {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CShopBuyResult), resp)
	}

	if req.ItemId == 0 || req.Count == 0 {
		resp.ErrCode = uint32(protocol.ErrorCode_Param_Invalid)
		return sendResp()
	}

	shopSys := GetShopSys(ctx)
	if shopSys == nil {
		resp.ErrCode = uint32(protocol.ErrorCode_Internal_Error)
		return sendResp()
	}
	if err := shopSys.Buy(ctx, req.ItemId, req.Count); err != nil {
		resp.ErrCode = errCodeFromError(err)
		return sendResp()
	}

	resp.ErrCode = uint32(protocol.ErrorCode_Success)
	if err := sendResp(); err != nil {
		return err
	}

	pushBagData(ctx, sessionId)
	pushMoneyData(ctx, sessionId)
	return nil
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysShop), func() iface.ISystem {
		return NewShopSys()
	})
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SShopBuy), handleShopBuy)
	})
}
