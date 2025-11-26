package entitysystem

import (
	"context"
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

const (
	defaultGoldMoneyID = uint32(protocol.MoneyType_MoneyTypeGoldCoin) // 使用MoneyTypeGoldCoin作为默认金币ID
	defaultGoldAmount  = 100000
)

// MoneySys 货币系统
type MoneySys struct {
	*BaseSystem
	moneyData *protocol.SiMoneyData
}

func NewMoneySys() *MoneySys {
	return &MoneySys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysMoney)),
	}
}

func GetMoneySys(ctx context.Context) *MoneySys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysMoney))
	if system == nil {
		return nil
	}
	moneySys, ok := system.(*MoneySys)
	if !ok || !moneySys.IsOpened() {
		return nil
	}
	return moneySys
}

func (ms *MoneySys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("money sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果money_data不存在，则初始化
	if binaryData.MoneyData == nil {
		binaryData.MoneyData = &protocol.SiMoneyData{
			MoneyMap: make(map[uint32]int64),
		}
	}
	ms.moneyData = binaryData.MoneyData

	// 如果MoneyMap为空，初始化默认金币
	if len(ms.moneyData.MoneyMap) == 0 {
		ms.moneyData.MoneyMap[defaultGoldMoneyID] = defaultGoldAmount
	}
}

func (ms *MoneySys) GetMoneyData() *protocol.SiMoneyData {
	return ms.moneyData
}

func (ms *MoneySys) GetAmount(moneyID uint32) int64 {
	if ms.moneyData == nil || ms.moneyData.MoneyMap == nil {
		return 0
	}
	return ms.moneyData.MoneyMap[moneyID]
}

// AddMoney 添加货币（通用处理器，特殊货币由特殊系统处理）
func (ms *MoneySys) AddMoney(ctx context.Context, moneyID uint32, amount int64) error {
	if amount == 0 {
		return nil
	}
	if amount < 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "amount must be positive")
	}

	// 特殊货币由特殊系统处理
	switch moneyID {
	case uint32(protocol.MoneyType_MoneyTypeExp):
		// 经验由等级系统处理
		levelSys := GetLevelSys(ctx)
		if levelSys == nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "level system not ready")
		}
		return levelSys.AddExp(ctx, uint64(amount))
	case uint32(protocol.MoneyType_MoneyTypeVipExp):
		// vip经验由VIP系统处理
		levelSys := GetVipSys(ctx)
		if levelSys == nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "vip system not ready")
		}
		return levelSys.AddVipExp(ctx, uint32(amount))
	case uint32(protocol.MoneyType_MoneyTypeActivePoint):
		activitySys := GetDailyActivitySys(ctx)
		if activitySys == nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "activity system not ready")
		}
		return activitySys.AddActivePoints(ctx, uint32(amount))
	default:
		// 普通货币由货币系统处理
		return ms.updateBalance(ctx, moneyID, amount)
	}
}

// SubMoney 扣除货币（通用处理器，特殊货币由特殊系统处理）
func (ms *MoneySys) SubMoney(ctx context.Context, moneyID uint32, amount int64) error {
	if amount <= 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "amount must be positive")
	}

	// 特殊货币由特殊系统处理
	switch moneyID {
	case uint32(protocol.MoneyType_MoneyTypeExp):
		// 经验由等级系统处理（通常经验不能扣除，但保留接口）
		levelSys := GetLevelSys(ctx)
		if levelSys == nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "level system not ready")
		}
		// 经验通常不能扣除，这里返回错误
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "experience cannot be deducted")
	case uint32(protocol.MoneyType_MoneyTypeActivePoint):
		activitySys := GetDailyActivitySys(ctx)
		if activitySys == nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "activity system not ready")
		}
		return activitySys.CostActivePoints(ctx, uint32(amount))
	default:
		// 普通货币由货币系统处理
		return ms.updateBalance(ctx, moneyID, -amount)
	}
}

// CostMoney 扣除货币（兼容旧接口，调用SubMoney）
func (ms *MoneySys) CostMoney(ctx context.Context, moneyID uint32, amount int64) error {
	return ms.SubMoney(ctx, moneyID, amount)
}

func (ms *MoneySys) updateBalance(ctx context.Context, moneyID uint32, delta int64) error {
	return ms.UpdateBalanceTx(ctx, moneyID, delta)
}

// UpdateBalanceTx 更新余额
func (ms *MoneySys) UpdateBalanceTx(ctx context.Context, moneyID uint32, delta int64) error {
	_, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	// 确保moneyData和MoneyMap已初始化
	if ms.moneyData == nil {
		playerRole, err := GetIPlayerRoleByContext(ctx)
		if err != nil {
			return err
		}
		binaryData := playerRole.GetBinaryData()
		if binaryData.MoneyData == nil {
			binaryData.MoneyData = &protocol.SiMoneyData{
				MoneyMap: make(map[uint32]int64),
			}
		}
		ms.moneyData = binaryData.MoneyData
	}
	if ms.moneyData.MoneyMap == nil {
		ms.moneyData.MoneyMap = make(map[uint32]int64)
	}

	current := ms.moneyData.MoneyMap[moneyID]
	newAmount := current + delta

	// 检查是否溢出（int64最大值）
	const maxInt64 = int64(^uint64(0) >> 1)
	if newAmount > maxInt64 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "money overflow: amount exceeds maximum")
	}

	if newAmount < 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "money not enough")
	}

	// 不再需要数据库操作，数据已存储在BinaryData中，会在OnLogout时统一保存
	ms.moneyData.MoneyMap[moneyID] = newAmount
	return nil
}

// UpdateBalanceOnlyMemory 仅更新内存状态（用于事务回滚后的恢复）
func (ms *MoneySys) UpdateBalanceOnlyMemory(moneyID uint32, amount int64) {
	if ms.moneyData == nil || ms.moneyData.MoneyMap == nil {
		return
	}
	ms.moneyData.MoneyMap[moneyID] = amount
}

func handleOpenMoney(ctx context.Context, _ *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	moneySys := GetMoneySys(ctx)
	var moneyData *protocol.SiMoneyData
	if moneySys != nil {
		moneyData = moneySys.GetMoneyData()
	} else {
		moneyData = &protocol.SiMoneyData{
			MoneyMap: map[uint32]int64{},
		}
	}
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CMoneyData), &protocol.S2CMoneyDataReq{
		MoneyData: moneyData,
	})
}

func pushMoneyData(ctx context.Context, sessionId string) {
	moneySys := GetMoneySys(ctx)
	if moneySys == nil {
		return
	}
	if err := gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CMoneyData), &protocol.S2CMoneyDataReq{
		MoneyData: moneySys.GetMoneyData(),
	}); err != nil {
		log.Errorf("push money data failed: %v", err)
	}
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysMoney), func() iface.ISystem {
		return NewMoneySys()
	})
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SOpenMoney), handleOpenMoney)
	})
}
