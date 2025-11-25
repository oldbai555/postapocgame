package entitysystem

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/dungeonserverlink"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/manager"
	"postapocgame/server/service/gameserver/internel/playeractor/clientprotocol"
	"time"
)

// FubenSys 副本系统
type FubenSys struct {
	*BaseSystem
	dungeonData *protocol.SiDungeonData
}

// NewFubenSys 创建副本系统
func NewFubenSys() *FubenSys {
	return &FubenSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysDungeon)),
	}
}

func GetFubenSys(ctx context.Context) *FubenSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysDungeon))
	if system == nil {
		return nil
	}
	fubenSys, ok := system.(*FubenSys)
	if !ok || !fubenSys.IsOpened() {
		return nil
	}
	return fubenSys
}

// OnInit 初始化时从PlayerRoleBinaryData加载副本数据
func (fs *FubenSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("fuben sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果dungeon_data不存在，则初始化
	if binaryData.DungeonData == nil {
		binaryData.DungeonData = &protocol.SiDungeonData{
			Records: make([]*protocol.DungeonRecord, 0),
		}
	}
	fs.dungeonData = binaryData.DungeonData

	// 如果Records为空，初始化为空切片
	if fs.dungeonData.Records == nil {
		fs.dungeonData.Records = make([]*protocol.DungeonRecord, 0)
	}

	log.Infof("FubenSys initialized: RecordCount=%d", len(fs.dungeonData.Records))
}

// GetDungeonRecord 获取副本记录
func (fs *FubenSys) GetDungeonRecord(dungeonID uint32, difficulty uint32) *protocol.DungeonRecord {
	if fs.dungeonData == nil || fs.dungeonData.Records == nil {
		return nil
	}
	for _, record := range fs.dungeonData.Records {
		if record != nil && record.DungeonId == dungeonID && record.Difficulty == difficulty {
			return record
		}
	}
	return nil
}

// GetOrCreateDungeonRecord 获取或创建副本记录
func (fs *FubenSys) GetOrCreateDungeonRecord(dungeonID uint32, difficulty uint32) *protocol.DungeonRecord {
	if fs.dungeonData == nil {
		return nil
	}
	if fs.dungeonData.Records == nil {
		fs.dungeonData.Records = make([]*protocol.DungeonRecord, 0)
	}

	// 查找现有记录
	for _, record := range fs.dungeonData.Records {
		if record != nil && record.DungeonId == dungeonID && record.Difficulty == difficulty {
			return record
		}
	}

	// 创建新记录
	now := servertime.Now().Unix()
	newRecord := &protocol.DungeonRecord{
		DungeonId:     dungeonID,
		Difficulty:    difficulty,
		LastEnterTime: now,
		EnterCount:    0,
		ResetTime:     now,
	}
	fs.dungeonData.Records = append(fs.dungeonData.Records, newRecord)
	return newRecord
}

// CheckDungeonCD 检查副本CD（冷却时间）
func (fs *FubenSys) CheckDungeonCD(dungeonID uint32, difficulty uint32, cdMinutes uint32) (bool, time.Duration) {
	record := fs.GetDungeonRecord(dungeonID, difficulty)
	if record == nil {
		// 没有记录，可以进入
		return true, 0
	}

	now := servertime.Now()
	lastEnterTime := time.Unix(record.LastEnterTime, 0)
	elapsed := now.Sub(lastEnterTime)
	cdDuration := time.Duration(cdMinutes) * time.Minute

	if elapsed < cdDuration {
		// 还在CD中
		remaining := cdDuration - elapsed
		return false, remaining
	}

	return true, 0
}

// EnterDungeon 进入副本（更新进入时间和次数）
func (fs *FubenSys) EnterDungeon(dungeonID uint32, difficulty uint32) error {
	if fs.dungeonData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "dungeon data not initialized")
	}

	record := fs.GetOrCreateDungeonRecord(dungeonID, difficulty)
	if record == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "failed to create dungeon record")
	}

	now := servertime.Now()
	lastResetTime := time.Unix(record.ResetTime, 0)

	// 检查是否需要重置（每日重置）
	if now.Sub(lastResetTime) >= 24*time.Hour {
		record.EnterCount = 1
		record.ResetTime = now.Unix()
	} else {
		record.EnterCount++
	}
	record.LastEnterTime = now.Unix()

	return nil
}

// GetDungeonData 获取副本数据（用于协议）
func (fs *FubenSys) GetDungeonData() *protocol.SiDungeonData {
	return fs.dungeonData
}

// GetAllRecords 获取所有副本记录
func (fs *FubenSys) GetAllRecords() []*protocol.DungeonRecord {
	if fs.dungeonData == nil || fs.dungeonData.Records == nil {
		return make([]*protocol.DungeonRecord, 0)
	}
	return fs.dungeonData.Records
}

// handleEnterDungeon 处理进入副本请求（限时副本）
func handleEnterDungeon(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	var req protocol.C2SEnterDungeonReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// 获取玩家角色
	roleMgr := manager.GetPlayerRoleManager()
	playerRole := roleMgr.GetBySession(sessionId)
	if playerRole == nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "玩家角色不存在",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 获取副本配置
	configMgr := jsonconf.GetConfigManager()
	dungeonCfg, ok := configMgr.GetDungeonConfig(req.DungeonId)
	if !ok || dungeonCfg == nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "副本不存在",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 检查是否为限时副本
	if dungeonCfg.Type != 2 {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "该副本不是限时副本",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 检查难度是否有效
	var difficultyCfg *jsonconf.DungeonDifficulty
	for i := range dungeonCfg.Difficulties {
		if dungeonCfg.Difficulties[i].Difficulty == req.Difficulty {
			difficultyCfg = dungeonCfg.Difficulties[i]
			break
		}
	}
	if difficultyCfg == nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "难度不存在",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 获取副本系统
	roleCtx := playerRole.WithContext(ctx)
	fubenSys := GetFubenSys(roleCtx)
	if fubenSys == nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "副本系统未初始化",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 检查每日进入次数
	record := fubenSys.GetDungeonRecord(req.DungeonId, req.Difficulty)
	if record != nil {
		now := servertime.Now()
		lastResetTime := time.Unix(record.ResetTime, 0)

		// 检查是否需要重置（每日重置）
		if now.Sub(lastResetTime) >= 24*time.Hour {
			record.EnterCount = 0
			record.ResetTime = now.Unix()
		}

		// 检查每日最大进入次数
		if dungeonCfg.MaxEnterPerDay > 0 && record.EnterCount >= dungeonCfg.MaxEnterPerDay {
			resp := &protocol.S2CEnterDungeonResultReq{
				Success:   false,
				Message:   "今日进入次数已用完",
				DungeonId: req.DungeonId,
			}
			return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
		}
	}

	// 检查消耗物品（如通天令）
	if len(difficultyCfg.ConsumeItems) > 0 {
		// 检查消耗是否足够
		if err := playerRole.CheckConsume(ctx, difficultyCfg.ConsumeItems); err != nil {
			resp := &protocol.S2CEnterDungeonResultReq{
				Success:   false,
				Message:   "消耗物品不足",
				DungeonId: req.DungeonId,
			}
			return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
		}
		// 扣除消耗（在Actor中执行）
		roleCtx := playerRole.WithContext(ctx)
		if err := playerRole.ApplyConsume(roleCtx, difficultyCfg.ConsumeItems); err != nil {
			log.Errorf("ApplyConsume failed: %v", err)
			resp := &protocol.S2CEnterDungeonResultReq{
				Success:   false,
				Message:   "扣除消耗失败",
				DungeonId: req.DungeonId,
			}
			return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
		}
	}

	// 更新进入记录
	if err := fubenSys.EnterDungeon(req.DungeonId, req.Difficulty); err != nil {
		log.Errorf("EnterDungeon failed: %v", err)
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "进入副本失败",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 获取角色信息
	roleInfo := playerRole.GetRoleInfo()
	if roleInfo == nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "角色信息不存在",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 汇总属性
	attrSys := GetAttrSys(roleCtx)
	var syncAttrData *protocol.SyncAttrData
	if attrSys != nil {
		allAttrs := attrSys.CalculateAllAttrs(roleCtx)
		if len(allAttrs) > 0 {
			syncAttrData = &protocol.SyncAttrData{
				AttrData: allAttrs,
			}
		}
	}

	// 获取技能列表
	skillSys := GetSkillSys(roleCtx)
	var skillMap map[uint32]uint32
	if skillSys != nil {
		skillMap = skillSys.GetSkillMap()
	} else {
		skillMap = make(map[uint32]uint32)
	}

	// 构造进入副本请求
	reqData, err := proto.Marshal(&protocol.G2DEnterDungeonReq{
		SessionId:    sessionId,
		PlatformId:   gshare.GetPlatformId(),
		SrvId:        gshare.GetSrvId(),
		SimpleData:   roleInfo,
		SyncAttrData: syncAttrData,
		SkillMap:     skillMap,
		DungeonId:    req.DungeonId,  // 传递副本ID
		Difficulty:   req.Difficulty, // 传递难度
	})
	if err != nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "系统错误",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 使用带SessionId的异步RPC调用
	srvType := uint8(protocol.SrvType_SrvTypeDungeonServer)
	err = dungeonserverlink.AsyncCall(context.Background(), srvType, sessionId, uint16(protocol.G2DRpcProtocol_G2DEnterDungeon), reqData)
	if err != nil {
		log.Errorf("call dungeon service enter dungeon failed: %v", err)
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "进入副本失败",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 发送成功响应
	resp := &protocol.S2CEnterDungeonResultReq{
		Success:   true,
		Message:   "进入副本成功",
		DungeonId: req.DungeonId,
	}
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
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
	fubenSys := GetFubenSys(roleCtx)
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
				levelSys := GetLevelSys(roleCtx)
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

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysDungeon), func() iface.ISystem {
		return NewFubenSys()
	})
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SEnterDungeon), handleEnterDungeon)
		// 注册副本结算的RPC处理器
		dungeonserverlink.RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GSettleDungeon), handleSettleDungeon)
		// 注册进入副本成功的RPC处理器
		dungeonserverlink.RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GEnterDungeonSuccess), handleEnterDungeonSuccess)
	})
}
