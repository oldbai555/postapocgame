package entity

import (
	"context"
	"postapocgame/server/internal"
	icalc "postapocgame/server/internal/attrcalc"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/adapter/gateway"
	gameattrcalc "postapocgame/server/service/gameserver/internel/adapter/system/attrcalc"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/attr"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// 确保 AttrCalculator 实现 IAttrCalculator 接口
var _ interfaces.IAttrCalculator = (*AttrCalculator)(nil)

// AttrCalculator 属性计算工具类（从 AttrSystemAdapter 重构而来）
// 不依赖 Actor 框架，只依赖接口（Repository、ConfigManager 等）
type AttrCalculator struct {
	attrDataMap           map[uint32]*protocol.AttrVec
	dirtySystems          map[uint32]bool
	calculators           map[uint32]gameattrcalc.Calculator
	addRateCalcs          map[uint32]gameattrcalc.AddRateCalculator
	sysAttr               map[uint32]*icalc.FightAttrCalc
	sysAddRateAttr        map[uint32]*icalc.FightAttrCalc
	extraAttr             *icalc.ExtraAttrCalc
	fightAttr             *icalc.FightAttrCalc
	extraDirty            map[attrdef.AttrType]struct{}
	sysPowerMap           map[uint32]int64
	sysAttrChanged        bool
	calcSysPowerUseCase   *attr.CalculateSysPowerUseCase
	compareAttrVecUseCase *attr.CompareAttrVecUseCase
	dungeonGateway        interfaces.DungeonServerGateway
	networkGateway        gateway.NetworkGateway
	playerRole            iface.IPlayerRole
}

// NewAttrCalculator 创建属性计算工具类
func NewAttrCalculator(playerRole iface.IPlayerRole) *AttrCalculator {
	container := di.GetContainer()
	calcSysPowerUC := attr.NewCalculateSysPowerUseCase(container.ConfigGateway())
	compareAttrVecUC := attr.NewCompareAttrVecUseCase()
	return &AttrCalculator{
		attrDataMap:           make(map[uint32]*protocol.AttrVec),
		dirtySystems:          make(map[uint32]bool),
		calculators:           make(map[uint32]gameattrcalc.Calculator),
		addRateCalcs:          make(map[uint32]gameattrcalc.AddRateCalculator),
		sysAttr:               make(map[uint32]*icalc.FightAttrCalc),
		sysAddRateAttr:        make(map[uint32]*icalc.FightAttrCalc),
		extraAttr:             icalc.NewExtraAttrCalc(),
		fightAttr:             icalc.NewFightAttrCalc(),
		extraDirty:            make(map[attrdef.AttrType]struct{}),
		sysPowerMap:           make(map[uint32]int64),
		calcSysPowerUseCase:   calcSysPowerUC,
		compareAttrVecUseCase: compareAttrVecUC,
		dungeonGateway:        container.DungeonServerGateway(),
		networkGateway:        container.GetNetworkGateway(),
		playerRole:            playerRole,
	}
}

// Init 初始化计算器（从上下文克隆）
func (ac *AttrCalculator) Init(ctx context.Context) {
	ac.cloneCalculators(ctx)
	ac.cloneAddRateCalculators(ctx)
}

// OnInit 实现 IAttrCalculator 接口（别名，调用 Init）
func (ac *AttrCalculator) OnInit(ctx context.Context) {
	ac.Init(ctx)
}

func (ac *AttrCalculator) cloneCalculators(ctx context.Context) {
	calculators := gameattrcalc.CloneCalculators(ctx)
	if len(calculators) == 0 {
		ac.calculators = make(map[uint32]gameattrcalc.Calculator)
		return
	}
	ac.calculators = calculators
}

func (ac *AttrCalculator) cloneAddRateCalculators(ctx context.Context) {
	calculators := gameattrcalc.CloneAddRateCalculators(ctx)
	if len(calculators) == 0 {
		ac.addRateCalcs = make(map[uint32]gameattrcalc.AddRateCalculator)
		return
	}
	ac.addRateCalcs = calculators
}

// MarkDirty 标记需要重算的系统
func (ac *AttrCalculator) MarkDirty(saAttrSysId uint32) {
	ac.dirtySystems[saAttrSysId] = true
	ac.sysAttrChanged = true
}

// ResetSysAttr 对外暴露的系统属性重算入口
func (ac *AttrCalculator) ResetSysAttr(saAttrSysId uint32) {
	ac.MarkDirty(saAttrSysId)
}

// CalculateAllAttrs 计算所有系统的属性
func (ac *AttrCalculator) CalculateAllAttrs(ctx context.Context) map[uint32]*protocol.AttrVec {
	result := make(map[uint32]*protocol.AttrVec)
	for sysID := range ac.calculators {
		if vec, exists := ac.attrDataMap[sysID]; exists && !ac.isSystemDirty(sysID) {
			result[sysID] = vec
			continue
		}
		attrVec, changed := ac.calculateSystemAttr(ctx, sysID)
		if attrVec == nil {
			continue
		}
		result[sysID] = attrVec
		if changed {
			ac.sysAttrChanged = true
		}
		ac.clearSystemDirty(sysID)
	}
	ac.dirtySystems = make(map[uint32]bool)
	return result
}

// RunOne 计算变动的系统属性并同步到DungeonServer
func (ac *AttrCalculator) RunOne(ctx context.Context) error {
	if len(ac.dirtySystems) == 0 {
		return nil
	}

	if ac.playerRole == nil {
		log.Errorf("RunOne: playerRole is nil")
		return nil
	}

	changedAttrs := make(map[uint32]*protocol.AttrVec)
	for saAttrSysId := range ac.dirtySystems {
		attrVec, changed := ac.calculateSystemAttr(ctx, saAttrSysId)
		ac.clearSystemDirty(saAttrSysId)
		if attrVec == nil || !changed {
			continue
		}
		changedAttrs[saAttrSysId] = attrVec
	}
	ac.dirtySystems = make(map[uint32]bool)
	if len(changedAttrs) == 0 {
		return nil
	}
	ac.sysAttrChanged = true
	ac.calcSysPowerMap(ctx)
	ac.syncAttrsToDungeonServer(ctx, ac.playerRole, changedAttrs)
	return nil
}

// PushFullAttrData 将当前缓存的属性完整推送给客户端
func (ac *AttrCalculator) PushFullAttrData(ctx context.Context) {
	if ac.playerRole == nil {
		log.Errorf("push full attr data failed: playerRole is nil")
		return
	}
	if len(ac.attrDataMap) == 0 {
		return
	}
	ac.calcSysPowerMap(ctx)
	syncData := &protocol.SyncAttrData{
		AttrData: cloneAttrDataMap(ac.attrDataMap),
	}
	if len(syncData.AttrData) == 0 {
		return
	}
	ac.pushAttrDataToClient(ctx, ac.playerRole, syncData)
}

// PushSyncDataToClient 按需推送指定的属性数据
func (ac *AttrCalculator) PushSyncDataToClient(ctx context.Context, syncData *protocol.SyncAttrData) {
	if syncData == nil {
		return
	}
	if ac.playerRole == nil {
		log.Errorf("push attr data failed: playerRole is nil")
		return
	}
	ac.calcSysPowerMap(ctx)
	ac.pushAttrDataToClient(ctx, ac.playerRole, syncData)
}

// syncAttrsToDungeonServer 同步属性到DungeonServer
func (ac *AttrCalculator) syncAttrsToDungeonServer(ctx context.Context, playerRole iface.IPlayerRole, changedAttrs map[uint32]*protocol.AttrVec) {
	ac.calcTotalSysAddRate(ctx)
	// 构建同步数据
	syncData := &protocol.SyncAttrData{
		AttrData: make(map[uint32]*protocol.AttrVec),
	}
	for saAttrSysId, attrVec := range changedAttrs {
		syncData.AttrData[saAttrSysId] = attrVec
	}
	if addRateVec := ac.buildAddRateAttrVec(); addRateVec != nil {
		syncData.AddRateAttr = addRateVec
	}
	ac.pushAttrDataToClient(ctx, playerRole, syncData)

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

	// 发送请求到DungeonActor（通过 DungeonActorMsgId 枚举）
	sessionID := playerRole.GetSessionId()
	err = ac.dungeonGateway.AsyncCall(ctx, sessionID, uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdSyncAttrs), reqData)
	if err != nil {
		log.Errorf("sync attrs to dungeon server failed: %v", err)
		return
	}

	log.Debugf("AttrCalculator synced attrs to DungeonServer: RoleId=%d, Systems=%d", playerRole.GetPlayerRoleId(), len(changedAttrs))
}

func (ac *AttrCalculator) calculateSystemAttr(ctx context.Context, sysID uint32) (*protocol.AttrVec, bool) {
	calculator, exists := ac.calculators[sysID]
	if !exists || calculator == nil {
		_, had := ac.attrDataMap[sysID]
		delete(ac.attrDataMap, sysID)
		delete(ac.sysAttr, sysID)
		return nil, had
	}
	rawAttrs := calculator.CalculateAttrs(ctx)
	if len(rawAttrs) == 0 {
		delete(ac.attrDataMap, sysID)
		delete(ac.sysAttr, sysID)
		return nil, true
	}
	attrVec := &protocol.AttrVec{
		Attrs: cloneAttrList(rawAttrs),
	}
	// 使用 UseCase 比较属性向量（纯业务逻辑已下沉）
	changed := !ac.compareAttrVecUseCase.Execute(ac.attrDataMap[sysID], attrVec)
	ac.attrDataMap[sysID] = attrVec
	ac.refreshSysCalc(sysID, attrVec)
	return attrVec, changed
}

func (ac *AttrCalculator) refreshSysCalc(sysID uint32, attrVec *protocol.AttrVec) {
	if attrVec == nil {
		return
	}
	calc := ac.ensureSysCalc(sysID)
	calc.Reset()
	for _, attr := range attrVec.Attrs {
		if attr == nil {
			continue
		}
		attrType := attrdef.AttrType(attr.Type)
		calc.AddValue(attrType, attrdef.AttrValue(attr.Value))
	}
}

func (ac *AttrCalculator) ensureSysCalc(sysID uint32) *icalc.FightAttrCalc {
	if calc, ok := ac.sysAttr[sysID]; ok && calc != nil {
		return calc
	}
	calc := icalc.NewFightAttrCalc()
	ac.sysAttr[sysID] = calc
	return calc
}

func (ac *AttrCalculator) ensureSysAddRateCalc(sysID uint32) *icalc.FightAttrCalc {
	if calc, ok := ac.sysAddRateAttr[sysID]; ok && calc != nil {
		return calc
	}
	calc := icalc.NewFightAttrCalc()
	ac.sysAddRateAttr[sysID] = calc
	return calc
}

func cloneAttrList(attrs []*protocol.AttrSt) []*protocol.AttrSt {
	if len(attrs) == 0 {
		return nil
	}
	cloned := make([]*protocol.AttrSt, 0, len(attrs))
	for _, attr := range attrs {
		if attr == nil {
			continue
		}
		cloned = append(cloned, &protocol.AttrSt{
			Type:  attr.Type,
			Value: attr.Value,
		})
	}
	return cloned
}

func cloneAttrDataMap(data map[uint32]*protocol.AttrVec) map[uint32]*protocol.AttrVec {
	if len(data) == 0 {
		return nil
	}
	cloned := make(map[uint32]*protocol.AttrVec, len(data))
	for sysID, vec := range data {
		if vec == nil {
			continue
		}
		cloned[sysID] = &protocol.AttrVec{
			Attrs: cloneAttrList(vec.Attrs),
		}
	}
	return cloned
}

// ApplyDungeonSyncData 同步战斗服回写的属性结果
func (ac *AttrCalculator) ApplyDungeonSyncData(syncData *protocol.SyncAttrData) {
	if syncData == nil || ac.fightAttr == nil {
		return
	}
	ac.fightAttr.Reset()
	for _, vec := range syncData.AttrData {
		if vec == nil || len(vec.Attrs) == 0 {
			continue
		}
		for _, attr := range vec.Attrs {
			if attr == nil {
				continue
			}
			attrType := attrdef.AttrType(attr.Type)
			if !attrdef.IsCombatAttr(attrType) {
				continue
			}
			ac.fightAttr.AddValue(attrType, attrdef.AttrValue(attr.Value))
		}
	}
	if addRate := syncData.AddRateAttr; addRate != nil {
		for _, attr := range addRate.Attrs {
			if attr == nil {
				continue
			}
			attrType := attrdef.AttrType(attr.Type)
			if !attrdef.IsCombatAttr(attrType) {
				continue
			}
			ac.fightAttr.AddValue(attrType, attrdef.AttrValue(attr.Value))
		}
	}
}

func (ac *AttrCalculator) calcSysPowerMap(ctx context.Context) {
	if len(ac.sysAttr) == 0 {
		ac.sysPowerMap = make(map[uint32]int64)
		return
	}
	var job uint32
	if ac.playerRole != nil {
		job = ac.playerRole.GetJob()
	}
	// 使用 UseCase 计算系统战力（纯业务逻辑已下沉）
	data := &attr.SystemAttrData{
		SysAttr:        ac.sysAttr,
		SysAddRateAttr: ac.sysAddRateAttr,
		Job:            job,
	}
	sysPower, err := ac.calcSysPowerUseCase.Execute(ctx, data)
	if err != nil {
		log.Errorf("calc sys power failed: %v", err)
		ac.sysPowerMap = make(map[uint32]int64)
		return
	}
	ac.sysPowerMap = sysPower
}

func (ac *AttrCalculator) isSystemDirty(sysID uint32) bool {
	_, ok := ac.dirtySystems[sysID]
	return ok
}

func (ac *AttrCalculator) clearSystemDirty(sysID uint32) {
	delete(ac.dirtySystems, sysID)
}

func (ac *AttrCalculator) buildAddRateAttrVec() *protocol.AttrVec {
	if len(ac.sysAddRateAttr) == 0 {
		return nil
	}
	attrs := make([]*protocol.AttrSt, 0)
	for _, calc := range ac.sysAddRateAttr {
		if calc == nil {
			continue
		}
		calc.DoRange(func(attrType attrdef.AttrType, value attrdef.AttrValue) {
			attrs = append(attrs, &protocol.AttrSt{
				Type:  uint32(attrType),
				Value: int64(value),
			})
		})
	}
	if len(attrs) == 0 {
		return nil
	}
	return &protocol.AttrVec{Attrs: attrs}
}

func (ac *AttrCalculator) calcTotalSysAddRate(ctx context.Context) {
	if len(ac.addRateCalcs) == 0 {
		for _, calc := range ac.sysAddRateAttr {
			if calc != nil {
				calc.Reset()
			}
		}
		return
	}
	totalCalc := icalc.NewFightAttrCalc()
	for _, calc := range ac.sysAttr {
		if calc == nil {
			continue
		}
		totalCalc.AddCalc(calc)
	}
	for sysID, calculator := range ac.addRateCalcs {
		if calculator == nil {
			continue
		}
		targetCalc := ac.ensureSysAddRateCalc(sysID)
		targetCalc.Reset()
		results := calculator.CalculateAddRate(ctx, totalCalc)
		for _, attr := range results {
			if attr == nil {
				continue
			}
			attrType := attrdef.AttrType(attr.Type)
			targetCalc.AddValue(attrType, attrdef.AttrValue(attr.Value))
		}
	}
}

func (ac *AttrCalculator) pushAttrDataToClient(ctx context.Context, playerRole iface.IPlayerRole, syncData *protocol.SyncAttrData) {
	if playerRole == nil || syncData == nil {
		return
	}
	cfg := getAttrPushConfig()
	filtered := ac.filterSyncAttrData(syncData, cfg)
	if filtered == nil {
		return
	}
	resp := &protocol.S2CAttrDataReq{
		AttrData: filtered,
	}
	if cfg.IncludeSysPowerMap {
		resp.SysPowerMap = ac.snapshotSysPowerMap()
	}
	sessionID := playerRole.GetSessionId()
	if err := ac.networkGateway.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CAttrData), resp); err != nil {
		log.Errorf("push attr data to client failed: %v", err)
	}
}

func getAttrPushConfig() *jsonconf.AttrPushConfig {
	return jsonconf.GetConfigManager().GetAttrPushConfig()
}

func shouldPushAttr(sysID uint32) bool {
	cfg := getAttrPushConfig()
	if cfg == nil {
		return false
	}
	if cfg.PushAll {
		return true
	}
	// 实时构建系统集合，支持热更新
	for _, id := range cfg.Systems {
		if id == sysID {
			return true
		}
	}
	return false
}

func (ac *AttrCalculator) filterSyncAttrData(data *protocol.SyncAttrData, cfg *jsonconf.AttrPushConfig) *protocol.SyncAttrData {
	if data == nil {
		return nil
	}
	result := &protocol.SyncAttrData{
		AttrData: make(map[uint32]*protocol.AttrVec),
	}
	for sysID, vec := range data.AttrData {
		if !shouldPushAttr(sysID) {
			continue
		}
		result.AttrData[sysID] = &protocol.AttrVec{
			Attrs: cloneAttrList(vec.Attrs),
		}
	}
	if data.AddRateAttr != nil {
		result.AddRateAttr = &protocol.AttrVec{
			Attrs: cloneAttrList(data.AddRateAttr.Attrs),
		}
	}
	if len(result.AttrData) == 0 && result.AddRateAttr == nil {
		return nil
	}
	return result
}

func (ac *AttrCalculator) snapshotSysPowerMap() map[uint32]int64 {
	if len(ac.sysPowerMap) == 0 {
		return nil
	}
	result := make(map[uint32]int64, len(ac.sysPowerMap))
	for k, v := range ac.sysPowerMap {
		result[k] = v
	}
	return result
}
