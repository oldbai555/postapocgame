package dungeonserverlink

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/manager"
	"postapocgame/server/service/gameserver/internel/playeractor/entitysystem"
)

// handleRegisterProtocols 处理DungeonServer注册协议的RPC请求
func handleRegisterProtocols(ctx context.Context, sessionId string, data []byte) error {
	var req protocol.D2GRegisterProtocolsReq
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal register protocols request failed: %v", err)
		return customerr.Wrap(err)
	}

	srvType := uint8(req.SrvType)
	log.Infof("received protocol registration from DungeonServer: srvType=%d, protocols=%d", srvType, len(req.Protocols))

	// 转换协议信息
	protocols := make([]struct {
		ProtoId  uint16
		IsCommon bool
	}, len(req.Protocols))

	for i, proto := range req.Protocols {
		protocols[i].ProtoId = uint16(proto.ProtoId)
		protocols[i].IsCommon = proto.IsCommon
		log.Debugf("  - protoId=%d, isCommon=%v", proto.ProtoId, proto.IsCommon)
	}

	// 注册到协议管理器
	if err := GetProtocolManager().RegisterProtocols(srvType, protocols); err != nil {
		log.Errorf("register protocols failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("successfully registered %d protocols for srvType=%d", len(protocols), srvType)
	return nil
}

// handleUnregisterProtocols 处理DungeonServer注销协议的RPC请求
func handleUnregisterProtocols(ctx context.Context, sessionId string, data []byte) error {
	var req protocol.D2GUnregisterProtocolsReq
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal unregister protocols request failed: %v", err)
		return customerr.Wrap(err)
	}

	srvType := uint8(req.SrvType)
	log.Infof("received protocol unregistration from DungeonServer: srvType=%d", srvType)

	// 从协议管理器注销
	if err := GetProtocolManager().UnregisterProtocols(srvType); err != nil {
		log.Errorf("unregister protocols failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("successfully unregistered protocols for srvType=%d", srvType)
	return nil
}

// handleSettleDungeon 处理副本结算的RPC请求
func handleSettleDungeon(ctx context.Context, sessionId string, data []byte) error {
	var req protocol.D2GSettleDungeonReq
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal settle dungeon request failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("received dungeon settlement: RoleId=%d, DungeonID=%d, Success=%v, Rewards=%d",
		req.RoleId, req.DungeonId, req.Success, len(req.Rewards))

	// 获取玩家角色
	playerRole := manager.GetPlayerRole(req.RoleId)
	if playerRole == nil {
		log.Errorf("player role not found: RoleId=%d", req.RoleId)
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found")
	}

	// 如果副本失败，只更新记录，不发放奖励
	if !req.Success {
		log.Infof("Dungeon failed, no rewards: RoleId=%d, DungeonID=%d", req.RoleId, req.DungeonId)
		return nil
	}

	// 更新副本记录
	roleCtx := playerRole.WithContext(ctx)
	fubenSys := entitysystem.GetFubenSys(roleCtx)
	if fubenSys != nil {
		if err := fubenSys.EnterDungeon(req.DungeonId, req.Difficulty); err != nil {
			log.Errorf("EnterDungeon failed: %v", err)
			// 不返回错误，继续发放奖励
		}
	}

	// 转换奖励格式并发放
	if len(req.Rewards) > 0 {
		rewards := make([]*jsonconf.ItemAmount, 0, len(req.Rewards))
		for _, reward := range req.Rewards {
			// 根据奖励类型转换
			var itemType uint32
			switch reward.Type {
			case 1: // 经验奖励
				// 经验奖励通过等级系统发放
				levelSys := entitysystem.GetLevelSys(roleCtx)
				if levelSys != nil {
					if err := levelSys.AddExp(roleCtx, uint64(reward.Count)); err != nil {
						log.Errorf("AddExp failed: %v", err)
					}
				}
				continue // 经验已处理，跳过
			case 2: // 金币奖励
				itemType = uint32(protocol.ItemType_ItemTypeMoney)
			case 3: // 物品奖励
				itemType = uint32(protocol.ItemType_ItemTypeMaterial)
			default:
				log.Warnf("Unknown reward type: %d", reward.Type)
				continue
			}

			rewards = append(rewards, &jsonconf.ItemAmount{
				ItemType: itemType,
				ItemId:   reward.ItemId,
				Count:    int64(reward.Count),
				Bind:     1, // 副本奖励默认绑定
			})
		}

		// 发放奖励
		if len(rewards) > 0 {
			if err := playerRole.GrantRewards(roleCtx, rewards); err != nil {
				log.Errorf("GrantRewards failed: %v", err)
				return customerr.Wrap(err)
			}
		}
	}

	log.Infof("Dungeon settled successfully: RoleId=%d, DungeonID=%d", req.RoleId, req.DungeonId)
	return nil
}

// handleEnterDungeonSuccess 处理进入副本成功通知
func handleEnterDungeonSuccess(ctx context.Context, sessionId string, data []byte) error {
	var req protocol.D2GEnterDungeonSuccessReq
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal enter dungeon success request failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("role entered dungeon successfully: RoleId=%d, SessionId=%s", req.RoleId, req.SessionId)
	// 这里可以添加后续处理逻辑，比如更新玩家状态等
	return nil
}

// handleAddItem 处理添加物品请求（拾取掉落物）- 在Actor中异步处理
func handleAddItem(ctx context.Context, sessionId string, data []byte) error {
	var req protocol.D2GAddItemReq
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal add item request failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("received add item request: RoleId=%d, ItemId=%d, Count=%d", req.RoleId, req.ItemId, req.Count)

	playerRole := manager.GetPlayerRole(req.RoleId)
	if playerRole == nil {
		log.Errorf("player role not found: RoleId=%d, SessionId=%s", req.RoleId, sessionId)
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found")
	}

	roleCtx := playerRole.WithContext(ctx)
	itemCfg, ok := jsonconf.GetConfigManager().GetItemConfig(req.ItemId)
	if !ok {
		log.Errorf("item config not found: ItemId=%d", req.ItemId)
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config not found")
	}

	rewards := []*jsonconf.ItemAmount{
		{
			ItemType: itemCfg.Type,
			ItemId:   req.ItemId,
			Count:    int64(req.Count),
			Bind:     0,
		},
	}

	if err := playerRole.GrantRewards(roleCtx, rewards); err != nil {
		log.Errorf("grant rewards failed: RoleId=%d, ItemId=%d, Count=%d, Error=%v", req.RoleId, req.ItemId, req.Count, err)
		_ = playerRole.SendProtoMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), &protocol.S2CPickupItemResultReq{
			Success: false,
			Message: "拾取失败，请稍后重试",
			ItemHdl: req.ItemHdl,
		})
		return customerr.Wrap(err)
	}

	if bagSys := entitysystem.GetBagSys(roleCtx); bagSys != nil {
		if err := playerRole.SendProtoMessage(uint16(protocol.S2CProtocol_S2CBagData), &protocol.S2CBagDataReq{
			BagData: bagSys.GetBagData(),
		}); err != nil {
			log.Warnf("send bag data failed: %v", err)
		}
	}

	log.Infof("item added successfully: RoleId=%d, ItemId=%d, Count=%d", req.RoleId, req.ItemId, req.Count)
	return nil
}

func handleAddItemActorMessage(message actor.IActorMessage) {
	ctx := message.GetContext()
	sessionId, _ := ctx.Value(gshare.ContextKeySession).(string)
	if err := handleAddItem(ctx, sessionId, message.GetData()); err != nil {
		log.Errorf("handleAddItem failed: %v", err)
	}
}

// handleSyncPosition 处理坐标同步的RPC请求
func handleSyncPosition(ctx context.Context, sessionId string, data []byte) error {
	var req protocol.D2GSyncPositionReq
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal sync position request failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Debugf("received position sync: RoleId=%d, SceneId=%d, Pos=(%d,%d)", req.RoleId, req.SceneId, req.PosX, req.PosY)

	// 获取玩家角色
	playerRole := manager.GetPlayerRole(req.RoleId)
	if playerRole == nil {
		log.Warnf("player role not found for position sync: RoleId=%d", req.RoleId)
		// 不返回错误，坐标同步失败不影响游戏流程
		return nil
	}

	// 更新角色坐标（如果需要存储的话）
	// 注意：当前GameServer不存储角色坐标，坐标由DungeonServer管理
	// 这里只是记录日志，如果需要可以扩展PlayerRole存储坐标信息

	log.Debugf("position synced: RoleId=%d, SceneId=%d, Pos=(%d,%d)", req.RoleId, req.SceneId, req.PosX, req.PosY)
	return nil
}

// InitProtocolRegistration 初始化协议注册相关的RPC处理器
func InitProtocolRegistration() {
	// 注册协议注册的RPC处理器
	RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GRegisterProtocols), handleRegisterProtocols)
	RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GUnregisterProtocols), handleUnregisterProtocols)
	// 注册副本结算的RPC处理器
	RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GSettleDungeon), handleSettleDungeon)
	// 注册进入副本成功的RPC处理器
	RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GEnterDungeonSuccess), handleEnterDungeonSuccess)
	// 注册添加物品的RPC处理器
	RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GAddItem), handleAddItem)
	gshare.RegisterHandler(uint16(protocol.D2GRpcProtocol_D2GAddItem), handleAddItemActorMessage)
	RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GSyncPosition), handleSyncPosition)

	log.Infof("protocol registration RPC handlers initialized")
}
