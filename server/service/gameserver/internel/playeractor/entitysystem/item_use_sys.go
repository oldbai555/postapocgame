package entitysystem

import (
	"context"
	"postapocgame/server/internal"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
	"time"
)

const (
	// 默认物品使用冷却时间（秒）
	DefaultItemUseCooldownSeconds int64 = 5
)

// ItemUseSys 物品使用系统（管理冷却时间等）
type ItemUseSys struct {
	*BaseSystem
	itemUseData *protocol.SiItemUseData
}

// NewItemUseSys 创建物品使用系统
func NewItemUseSys() *ItemUseSys {
	return &ItemUseSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysItemUse)),
	}
}

func GetItemUseSys(ctx context.Context) *ItemUseSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysItemUse))
	if system == nil {
		log.Errorf("not found system [%v] error:%v", protocol.SystemId_SysItemUse, err)
		return nil
	}
	sys := system.(*ItemUseSys)
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error:%v", protocol.SystemId_SysItemUse, err)
		return nil
	}
	return sys
}

// OnInit 初始化时从数据库加载数据
func (ius *ItemUseSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果item_use_data不存在，则初始化
	if binaryData.ItemUseData == nil {
		binaryData.ItemUseData = &protocol.SiItemUseData{
			CooldownMap: make(map[uint32]int64),
		}
	}
	ius.itemUseData = binaryData.ItemUseData

	log.Infof("ItemUseSys initialized")
}

// UseItem 使用物品
func (ius *ItemUseSys) UseItem(ctx context.Context, itemID uint32, count uint32) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 检查物品配置
	itemConfig, ok := jsonconf.GetConfigManager().GetItemConfig(itemID)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config not found: %d", itemID)
	}

	// 检查物品是否可使用（通过Flag检查）
	if itemConfig.Flag&uint64(protocol.ItemFlag_ItemFlagCanUse) == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item cannot be used")
	}

	// 检查物品类型（只有消耗品可以使用）
	if itemConfig.Type != uint32(protocol.ItemType_ItemTypeConsume) {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "only consume items can be used")
	}

	// 检查冷却时间
	now := time.Now().Unix()
	if ius.itemUseData == nil || ius.itemUseData.CooldownMap == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item use data not initialized")
	}
	if cooldownEnd, exists := ius.itemUseData.CooldownMap[itemID]; exists && cooldownEnd > now {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item is in cooldown")
	}

	// 检查背包中是否有该物品
	bagSys := GetBagSys(ctx)
	if bagSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag system not found")
	}

	if !bagSys.HasItem(itemID, count) {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not enough")
	}

	// 获取物品使用效果配置
	useEffectConfig, ok := jsonconf.GetConfigManager().GetItemUseEffectConfig(itemID)
	if !ok || useEffectConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item use effect config not found: %d", itemID)
	}

	// 应用物品效果
	var hpDelta int64 = 0
	var mpDelta int64 = 0
	var expDelta int64 = 0

	for i := uint32(0); i < count; i++ {
		// 遍历效果值数组，应用每个效果
		for _, value := range useEffectConfig.Values {
			switch useEffectConfig.EffectType {
			case 1: // 恢复HP
				hpDelta += int64(value)
			case 2: // 恢复MP
				mpDelta += int64(value)
			case 3: // 增加经验
				expDelta += int64(value)
			}
		}
	}

	// 如果有HP/MP变化，需要同步到DungeonServer
	if hpDelta != 0 || mpDelta != 0 {
		sessionId := playerRole.GetSessionId()
		if sessionId != "" {
			roleID := playerRole.GetPlayerRoleId()

			// 构造RPC请求
			reqData, err := internal.Marshal(&protocol.G2DUpdateHpMpReq{
				SessionId: sessionId,
				RoleId:    roleID,
				HpDelta:   hpDelta,
				MpDelta:   mpDelta,
			})
			if err != nil {
				log.Errorf("marshal update hp/mp request failed: %v", err)
			} else {
				// 异步调用DungeonServer更新HP/MP（通过IPlayerRole接口，避免循环依赖）
				err = playerRole.CallDungeonServer(ctx, uint16(protocol.G2DRpcProtocol_G2DUpdateHpMp), reqData)
				if err != nil {
					log.Errorf("call dungeon server update hp/mp failed: %v", err)
					// 不返回错误，继续执行
				}
			}
		}
	}

	// 如果有经验变化，通过等级系统添加经验
	if expDelta > 0 {
		levelSys := GetLevelSys(ctx)
		if levelSys != nil {
			err := levelSys.AddExp(ctx, uint64(expDelta))
			if err != nil {
				log.Errorf("add exp failed: %v", err)
			}
		}
	}

	// 扣除物品数量
	err = bagSys.RemoveItem(ctx, itemID, count)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 设置冷却时间（默认5秒，可以根据配置调整）
	cooldownSeconds := DefaultItemUseCooldownSeconds
	if ius.itemUseData.CooldownMap == nil {
		ius.itemUseData.CooldownMap = make(map[uint32]int64)
	}
	ius.itemUseData.CooldownMap[itemID] = now + cooldownSeconds

	log.Infof("Item used: ItemID=%d, Count=%d, HPDelta=%d, MPDelta=%d, ExpDelta=%d", itemID, count, hpDelta, mpDelta, expDelta)

	return nil
}

// CheckCooldown 检查物品是否在冷却中
func (ius *ItemUseSys) CheckCooldown(itemID uint32) bool {
	if ius.itemUseData == nil || ius.itemUseData.CooldownMap == nil {
		return false
	}
	now := time.Now().Unix()
	if cooldownEnd, exists := ius.itemUseData.CooldownMap[itemID]; exists && cooldownEnd > now {
		return true
	}
	return false
}

// GetCooldownRemaining 获取剩余冷却时间（秒）
func (ius *ItemUseSys) GetCooldownRemaining(itemID uint32) int64 {
	if ius.itemUseData == nil || ius.itemUseData.CooldownMap == nil {
		return 0
	}
	now := time.Now().Unix()
	if cooldownEnd, exists := ius.itemUseData.CooldownMap[itemID]; exists && cooldownEnd > now {
		return cooldownEnd - now
	}
	return 0
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysItemUse), func() iface.ISystem {
		return NewItemUseSys()
	})
}
