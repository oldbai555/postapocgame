package dungeonserverlink

import (
	"context"
	"postapocgame/server/internal"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/manager"
	"postapocgame/server/service/gameserver/internel/playeractor/entitysystem"
)

// handleRegisterProtocols 处理DungeonServer注册协议的RPC请求
func handleRegisterProtocols(ctx context.Context, sessionId string, data []byte) error {
	var req protocol.D2GRegisterProtocolsReq
	if err := internal.Unmarshal(data, &req); err != nil {
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
	if err := internal.Unmarshal(data, &req); err != nil {
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
	if err := internal.Unmarshal(data, &req); err != nil {
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
	if err := internal.Unmarshal(data, &req); err != nil {
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
	if err := internal.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal add item request failed: %v", err)
		sendAddItemResponse(ctx, false, "解析请求失败", req.ItemHdl)
		return nil
	}

	log.Infof("received add item request: RoleId=%d, ItemId=%d, Count=%d", req.RoleId, req.ItemId, req.Count)

	// 获取玩家角色
	playerRole := manager.GetPlayerRole(req.RoleId)
	if playerRole == nil {
		log.Errorf("player role not found: RoleId=%d", req.RoleId)
		sendAddItemResponse(ctx, false, "玩家角色不存在", req.ItemHdl)
		return nil
	}

	// 添加物品到背包
	roleCtx := playerRole.WithContext(ctx)
	bagSys := entitysystem.GetBagSys(roleCtx)
	if bagSys == nil {
		log.Errorf("bag system not found: RoleId=%d", req.RoleId)
		sendAddItemResponse(ctx, false, "背包系统未初始化", req.ItemHdl)
		return nil
	}

	// 添加物品
	err := bagSys.AddItem(roleCtx, req.ItemId, req.Count, 0) // 0表示不绑定
	if err != nil {
		log.Errorf("add item failed: RoleId=%d, ItemId=%d, Count=%d, Error=%v", req.RoleId, req.ItemId, req.Count, err)
		sendAddItemResponse(ctx, false, err.Error(), req.ItemHdl)
		return nil
	}

	log.Infof("item added successfully: RoleId=%d, ItemId=%d, Count=%d", req.RoleId, req.ItemId, req.Count)
	sendAddItemResponse(ctx, true, "", req.ItemHdl)
	return nil
}

// sendAddItemResponse 发送添加物品的RPC响应
func sendAddItemResponse(ctx context.Context, success bool, errorMsg string, itemHdl uint64) {
	// 从context中获取RequestId和连接信息
	requestId, ok := ctx.Value("rpcRequestId").(uint32)
	if !ok {
		log.Errorf("rpcRequestId not found in context")
		return
	}

	conn, ok := ctx.Value("rpcConn").(network.IConnection)
	if !ok {
		log.Errorf("rpcConn not found in context")
		return
	}

	// 构造响应
	resp := &protocol.D2GAddItemResp{
		Success:  success,
		ErrorMsg: errorMsg,
		ItemHdl:  itemHdl, // 传递掉落物句柄
	}
	respData, err := internal.Marshal(resp)
	if err != nil {
		log.Errorf("marshal D2GAddItemResp failed: %v", err)
		return
	}

	// 发送RPC响应
	rpcResp := &network.RPCResponse{
		RequestId: requestId,
		Success:   success,
		Data:      respData,
	}

	// 获取codec并编码响应
	codec := network.DefaultCodec()
	respDataEncoded := codec.EncodeRPCResponse(rpcResp)
	defer network.PutBuffer(respDataEncoded)

	msg := network.GetMessage()
	defer network.PutMessage(msg)
	msg.Type = network.MsgTypeRPCResponse
	msg.Payload = respDataEncoded

	if err := conn.SendMessage(msg); err != nil {
		log.Errorf("send RPC response failed: %v", err)
	}
}

// handleSyncPosition 处理坐标同步的RPC请求
func handleSyncPosition(ctx context.Context, sessionId string, data []byte) error {
	var req protocol.D2GSyncPositionReq
	if err := internal.Unmarshal(data, &req); err != nil {
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
	RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GSyncPosition), handleSyncPosition)

	log.Infof("protocol registration RPC handlers initialized")
}
