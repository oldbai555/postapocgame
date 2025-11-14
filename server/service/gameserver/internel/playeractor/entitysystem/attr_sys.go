package entitysystem

import (
	"context"
	"postapocgame/server/internal"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
)

// IAttrCalculator 属性计算器接口
type IAttrCalculator interface {
	// CalculateAttrs 计算该系统的属性，返回属性列表
	CalculateAttrs(ctx context.Context) []*protocol.AttrSt
}

// AttrSys 属性系统汇总
type AttrSys struct {
	*BaseSystem
	// 属性数据：key为SaAttrSys枚举值，value为该系统的属性列表
	attrDataMap map[uint32]*protocol.AttrVec
	// 脏标记：标记需要重算的系统
	dirtySystems map[uint32]bool
	// 属性计算器注册表：key为SaAttrSys枚举值，value为属性计算器
	calculators map[uint32]IAttrCalculator
}

// NewAttrSys 创建属性系统
func NewAttrSys() *AttrSys {
	return &AttrSys{
		BaseSystem:   NewBaseSystem(uint32(protocol.SystemId_SysAttr)),
		attrDataMap:  make(map[uint32]*protocol.AttrVec),
		dirtySystems: make(map[uint32]bool),
		calculators:  make(map[uint32]IAttrCalculator),
	}
}

func GetAttrSys(ctx context.Context) *AttrSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysAttr))
	if system == nil {
		return nil
	}
	attrSys, ok := system.(*AttrSys)
	if !ok || !attrSys.IsOpened() {
		return nil
	}
	return attrSys
}

// OnInit 系统初始化
func (as *AttrSys) OnInit(ctx context.Context) {
	// 注册属性计算器
	as.registerCalculators(ctx)
}

// registerCalculators 注册所有属性计算器
func (as *AttrSys) registerCalculators(ctx context.Context) {
	// 注册等级系统属性计算器
	levelSys := GetLevelSys(ctx)
	if levelSys != nil {
		as.calculators[uint32(protocol.SaAttrSys_SaLevel)] = levelSys
	}

	// 注册装备系统属性计算器
	equipSys := GetEquipSys(ctx)
	if equipSys != nil {
		as.calculators[uint32(protocol.SaAttrSys_SaEquip)] = equipSys
	}

	log.Infof("AttrSys registered %d calculators", len(as.calculators))
}

// RegisterAttrCalculator 注册属性计算器（供外部系统调用）
func (as *AttrSys) RegisterAttrCalculator(saAttrSysId uint32, calculator IAttrCalculator) {
	as.calculators[saAttrSysId] = calculator
	log.Infof("AttrSys registered calculator for SaAttrSys=%d", saAttrSysId)
}

// MarkDirty 标记需要重算的系统
func (as *AttrSys) MarkDirty(saAttrSysId uint32) {
	as.dirtySystems[saAttrSysId] = true
	log.Debugf("AttrSys marked dirty: SaAttrSys=%d", saAttrSysId)
}

// CalculateAllAttrs 计算所有系统的属性（首次登录时调用）
// 优化：只计算未计算过的系统，避免重复计算
func (as *AttrSys) CalculateAllAttrs(ctx context.Context) map[uint32]*protocol.AttrVec {
	result := make(map[uint32]*protocol.AttrVec)

	// 遍历所有已注册的属性计算器
	for saAttrSysId, calculator := range as.calculators {
		if calculator == nil {
			continue
		}

		// 优化：如果已经计算过且没有标记为dirty，跳过计算
		if existingAttrVec, exists := as.attrDataMap[saAttrSysId]; exists {
			if !as.dirtySystems[saAttrSysId] {
				// 使用已计算的结果，避免重复计算
				result[saAttrSysId] = existingAttrVec
				continue
			}
		}

		// 计算该系统的属性
		attrs := calculator.CalculateAttrs(ctx)
		if len(attrs) > 0 {
			attrVec := &protocol.AttrVec{
				Attrs: attrs,
			}
			result[saAttrSysId] = attrVec
			as.attrDataMap[saAttrSysId] = attrVec
		} else {
			// 如果属性为空，删除该系统的属性数据
			delete(as.attrDataMap, saAttrSysId)
		}
	}

	// 清空所有脏标记（首次计算后，所有系统都是干净的）
	as.dirtySystems = make(map[uint32]bool)

	log.Infof("AttrSys calculated all attrs: %d systems", len(result))
	return result
}

// RunOne 计算变动的系统属性并同步到DungeonServer
func (as *AttrSys) RunOne(ctx context.Context) {
	if len(as.dirtySystems) == 0 {
		return
	}

	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("RunOne get role err:%v", err)
		return
	}

	// 计算变动的系统属性
	changedAttrs := make(map[uint32]*protocol.AttrVec)
	for saAttrSysId := range as.dirtySystems {
		calculator, exists := as.calculators[saAttrSysId]
		if !exists {
			continue
		}

		// 计算该系统的属性
		attrs := calculator.CalculateAttrs(ctx)
		if len(attrs) > 0 {
			attrVec := &protocol.AttrVec{
				Attrs: attrs,
			}
			changedAttrs[saAttrSysId] = attrVec
			as.attrDataMap[saAttrSysId] = attrVec
		} else {
			// 如果属性为空，删除该系统的属性数据
			delete(as.attrDataMap, saAttrSysId)
		}
	}

	// 清空脏标记
	as.dirtySystems = make(map[uint32]bool)

	// 如果有变动的属性，同步到DungeonServer
	if len(changedAttrs) > 0 {
		as.syncAttrsToDungeonServer(ctx, playerRole, changedAttrs)
	}
}

// syncAttrsToDungeonServer 同步属性到DungeonServer
func (as *AttrSys) syncAttrsToDungeonServer(ctx context.Context, playerRole iface.IPlayerRole, changedAttrs map[uint32]*protocol.AttrVec) {
	// 构建同步数据
	syncData := &protocol.SyncAttrData{
		AttrData: make(map[uint32]*protocol.AttrVec),
	}
	for saAttrSysId, attrVec := range changedAttrs {
		syncData.AttrData[saAttrSysId] = attrVec
	}

	// 构造RPC请求
	reqData, err := internal.Marshal(&protocol.G2DSyncAttrsReq{
		SessionId: playerRole.GetSessionId(),
		RoleId:    playerRole.GetPlayerRoleId(),
		SyncData:  syncData,
	})
	if err != nil {
		log.Errorf("marshal sync attrs request failed: %v", err)
		return
	}

	// 发送RPC请求到DungeonServer（通过IPlayerRole接口，避免循环依赖）
	err = playerRole.CallDungeonServer(ctx, uint16(protocol.G2DRpcProtocol_G2DSyncAttrs), reqData)
	if err != nil {
		log.Errorf("sync attrs to dungeon server failed: %v", err)
		return
	}

	log.Debugf("AttrSys synced attrs to DungeonServer: RoleId=%d, Systems=%d", playerRole.GetPlayerRoleId(), len(changedAttrs))
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysAttr), func() iface.ISystem {
		return NewAttrSys()
	})
}
