package system

import (
	"context"
	"postapocgame/server/internal"
	icalc "postapocgame/server/internal/attrcalc"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/attrpower"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/gateway"
	gameattrcalc "postapocgame/server/service/gameserver/internel/adapter/system/attrcalc"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// AttrSystemAdapter 属性系统适配器
type AttrSystemAdapter struct {
	*BaseSystemAdapter
	attrDataMap    map[uint32]*protocol.AttrVec
	dirtySystems   map[uint32]bool
	calculators    map[uint32]gameattrcalc.Calculator
	addRateCalcs   map[uint32]gameattrcalc.AddRateCalculator
	sysAttr        map[uint32]*icalc.FightAttrCalc
	sysAddRateAttr map[uint32]*icalc.FightAttrCalc
	extraAttr      *icalc.ExtraAttrCalc
	fightAttr      *icalc.FightAttrCalc
	extraDirty     map[attrdef.AttrType]struct{}
	sysPowerMap    map[uint32]int64
	sysAttrChanged bool
	dungeonGateway interfaces.DungeonServerGateway
	networkGateway gateway.NetworkGateway
}

// NewAttrSystemAdapter 创建属性系统适配器
func NewAttrSystemAdapter() *AttrSystemAdapter {
	container := di.GetContainer()
	return &AttrSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysAttr)),
		attrDataMap:       make(map[uint32]*protocol.AttrVec),
		dirtySystems:      make(map[uint32]bool),
		calculators:       make(map[uint32]gameattrcalc.Calculator),
		addRateCalcs:      make(map[uint32]gameattrcalc.AddRateCalculator),
		sysAttr:           make(map[uint32]*icalc.FightAttrCalc),
		sysAddRateAttr:    make(map[uint32]*icalc.FightAttrCalc),
		extraAttr:         icalc.NewExtraAttrCalc(),
		fightAttr:         icalc.NewFightAttrCalc(),
		extraDirty:        make(map[attrdef.AttrType]struct{}),
		sysPowerMap:       make(map[uint32]int64),
		dungeonGateway:    container.DungeonServerGateway(),
		networkGateway:    container.GetNetworkGateway(),
	}
}

// OnInit 系统初始化
func (a *AttrSystemAdapter) OnInit(ctx context.Context) {
	a.cloneCalculators(ctx)
	a.cloneAddRateCalculators(ctx)
}

func (a *AttrSystemAdapter) cloneCalculators(ctx context.Context) {
	calculators := gameattrcalc.CloneCalculators(ctx)
	if len(calculators) == 0 {
		a.calculators = make(map[uint32]gameattrcalc.Calculator)
		return
	}
	a.calculators = calculators
}

func (a *AttrSystemAdapter) cloneAddRateCalculators(ctx context.Context) {
	calculators := gameattrcalc.CloneAddRateCalculators(ctx)
	if len(calculators) == 0 {
		a.addRateCalcs = make(map[uint32]gameattrcalc.AddRateCalculator)
		return
	}
	a.addRateCalcs = calculators
}

// MarkDirty 标记需要重算的系统
func (a *AttrSystemAdapter) MarkDirty(saAttrSysId uint32) {
	a.dirtySystems[saAttrSysId] = true
	a.sysAttrChanged = true
}

// ResetSysAttr 对外暴露的系统属性重算入口
func (a *AttrSystemAdapter) ResetSysAttr(saAttrSysId uint32) {
	a.MarkDirty(saAttrSysId)
}

// CalculateAllAttrs 计算所有系统的属性
func (a *AttrSystemAdapter) CalculateAllAttrs(ctx context.Context) map[uint32]*protocol.AttrVec {
	result := make(map[uint32]*protocol.AttrVec)
	for sysID := range a.calculators {
		if vec, exists := a.attrDataMap[sysID]; exists && !a.isSystemDirty(sysID) {
			result[sysID] = vec
			continue
		}
		attrVec, changed := a.calculateSystemAttr(ctx, sysID)
		if attrVec == nil {
			continue
		}
		result[sysID] = attrVec
		if changed {
			a.sysAttrChanged = true
		}
		a.clearSystemDirty(sysID)
	}
	a.dirtySystems = make(map[uint32]bool)
	return result
}

// RunOne 计算变动的系统属性并同步到DungeonServer（实现 RunOneUseCase 接口）
func (a *AttrSystemAdapter) RunOne(ctx context.Context) error {
	if len(a.dirtySystems) == 0 {
		return nil
	}

	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("RunOne get role err:%v", err)
		return err
	}

	changedAttrs := make(map[uint32]*protocol.AttrVec)
	for saAttrSysId := range a.dirtySystems {
		attrVec, changed := a.calculateSystemAttr(ctx, saAttrSysId)
		a.clearSystemDirty(saAttrSysId)
		if attrVec == nil || !changed {
			continue
		}
		changedAttrs[saAttrSysId] = attrVec
	}
	a.dirtySystems = make(map[uint32]bool)
	if len(changedAttrs) == 0 {
		return nil
	}
	a.sysAttrChanged = true
	a.calcSysPowerMap(ctx)
	a.syncAttrsToDungeonServer(ctx, playerRole, changedAttrs)
	return nil
}

// PushFullAttrData 将当前缓存的属性完整推送给客户端（兼容旧 AttrSys 接口）
func (a *AttrSystemAdapter) PushFullAttrData(ctx context.Context) {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("push full attr data failed: %v", err)
		return
	}
	if len(a.attrDataMap) == 0 {
		return
	}
	a.calcSysPowerMap(ctx)
	syncData := &protocol.SyncAttrData{
		AttrData: cloneAttrDataMap(a.attrDataMap),
	}
	if len(syncData.AttrData) == 0 {
		return
	}
	a.pushAttrDataToClient(ctx, playerRole, syncData)
}

// PushSyncDataToClient 按需推送指定的属性数据（兼容旧 AttrSys 接口）
func (a *AttrSystemAdapter) PushSyncDataToClient(ctx context.Context, syncData *protocol.SyncAttrData) {
	if syncData == nil {
		return
	}
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("push attr data failed: %v", err)
		return
	}
	a.calcSysPowerMap(ctx)
	a.pushAttrDataToClient(ctx, playerRole, syncData)
}

// syncAttrsToDungeonServer 同步属性到DungeonServer
func (a *AttrSystemAdapter) syncAttrsToDungeonServer(ctx context.Context, playerRole iface.IPlayerRole, changedAttrs map[uint32]*protocol.AttrVec) {
	a.calcTotalSysAddRate(ctx)
	// 构建同步数据
	syncData := &protocol.SyncAttrData{
		AttrData: make(map[uint32]*protocol.AttrVec),
	}
	for saAttrSysId, attrVec := range changedAttrs {
		syncData.AttrData[saAttrSysId] = attrVec
	}
	if addRateVec := a.buildAddRateAttrVec(); addRateVec != nil {
		syncData.AddRateAttr = addRateVec
	}
	a.pushAttrDataToClient(ctx, playerRole, syncData)

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

	// 发送RPC请求到DungeonServer
	sessionID := playerRole.GetSessionId()
	srvType, _, ok := a.dungeonGateway.GetSrvTypeForProtocol(uint16(protocol.G2DRpcProtocol_G2DSyncAttrs))
	if !ok {
		log.Errorf("get srv type for protocol failed: G2DSyncAttrs")
		return
	}
	err = a.dungeonGateway.AsyncCall(ctx, srvType, sessionID, uint16(protocol.G2DRpcProtocol_G2DSyncAttrs), reqData)
	if err != nil {
		log.Errorf("sync attrs to dungeon server failed: %v", err)
		return
	}

	log.Debugf("AttrSys synced attrs to DungeonServer: RoleId=%d, Systems=%d", playerRole.GetPlayerRoleId(), len(changedAttrs))
}

func (a *AttrSystemAdapter) calculateSystemAttr(ctx context.Context, sysID uint32) (*protocol.AttrVec, bool) {
	calculator, exists := a.calculators[sysID]
	if !exists || calculator == nil {
		_, had := a.attrDataMap[sysID]
		delete(a.attrDataMap, sysID)
		delete(a.sysAttr, sysID)
		return nil, had
	}
	rawAttrs := calculator.CalculateAttrs(ctx)
	if len(rawAttrs) == 0 {
		delete(a.attrDataMap, sysID)
		delete(a.sysAttr, sysID)
		return nil, true
	}
	attrVec := &protocol.AttrVec{
		Attrs: cloneAttrList(rawAttrs),
	}
	changed := !attrVecEquals(a.attrDataMap[sysID], attrVec)
	a.attrDataMap[sysID] = attrVec
	a.refreshSysCalc(sysID, attrVec)
	return attrVec, changed
}

func (a *AttrSystemAdapter) refreshSysCalc(sysID uint32, attrVec *protocol.AttrVec) {
	if attrVec == nil {
		return
	}
	calc := a.ensureSysCalc(sysID)
	calc.Reset()
	for _, attr := range attrVec.Attrs {
		if attr == nil {
			continue
		}
		attrType := attrdef.AttrType(attr.Type)
		calc.AddValue(attrType, attrdef.AttrValue(attr.Value))
	}
}

func (a *AttrSystemAdapter) ensureSysCalc(sysID uint32) *icalc.FightAttrCalc {
	if calc, ok := a.sysAttr[sysID]; ok && calc != nil {
		return calc
	}
	calc := icalc.NewFightAttrCalc()
	a.sysAttr[sysID] = calc
	return calc
}

func (a *AttrSystemAdapter) ensureSysAddRateCalc(sysID uint32) *icalc.FightAttrCalc {
	if calc, ok := a.sysAddRateAttr[sysID]; ok && calc != nil {
		return calc
	}
	calc := icalc.NewFightAttrCalc()
	a.sysAddRateAttr[sysID] = calc
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

func attrVecEquals(a, b *protocol.AttrVec) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a.Attrs) != len(b.Attrs) {
		return false
	}
	tmp := make(map[uint32]int64, len(a.Attrs))
	for _, attr := range a.Attrs {
		if attr == nil {
			continue
		}
		tmp[attr.Type] = attr.Value
	}
	for _, attr := range b.Attrs {
		if attr == nil {
			continue
		}
		value, ok := tmp[attr.Type]
		if !ok || value != attr.Value {
			return false
		}
		delete(tmp, attr.Type)
	}
	return len(tmp) == 0
}

// ApplyDungeonSyncData 同步战斗服回写的属性结果
func (a *AttrSystemAdapter) ApplyDungeonSyncData(syncData *protocol.SyncAttrData) {
	if syncData == nil || a.fightAttr == nil {
		return
	}
	a.fightAttr.Reset()
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
			a.fightAttr.AddValue(attrType, attrdef.AttrValue(attr.Value))
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
			a.fightAttr.AddValue(attrType, attrdef.AttrValue(attr.Value))
		}
	}
}

func (a *AttrSystemAdapter) calcSysPowerMap(ctx context.Context) {
	if len(a.sysAttr) == 0 {
		a.sysPowerMap = make(map[uint32]int64)
		return
	}
	var job uint32
	if playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx); err == nil && playerRole != nil {
		job = playerRole.GetJob()
	}
	sysPower := make(map[uint32]int64, len(a.sysAttr))
	for sysID, calc := range a.sysAttr {
		if calc == nil {
			continue
		}
		temp := icalc.NewFightAttrCalc()
		temp.AddCalc(calc)
		if addRateCalc := a.sysAddRateAttr[sysID]; addRateCalc != nil {
			temp.AddCalc(addRateCalc)
		}
		icalc.ApplyConversions(temp)
		icalc.ApplyPercentages(temp)
		sysPower[sysID] = attrpower.CalcPower(temp, job)
	}
	a.sysPowerMap = sysPower
}

func (a *AttrSystemAdapter) isSystemDirty(sysID uint32) bool {
	_, ok := a.dirtySystems[sysID]
	return ok
}

func (a *AttrSystemAdapter) clearSystemDirty(sysID uint32) {
	delete(a.dirtySystems, sysID)
}

func (a *AttrSystemAdapter) buildAddRateAttrVec() *protocol.AttrVec {
	if len(a.sysAddRateAttr) == 0 {
		return nil
	}
	attrs := make([]*protocol.AttrSt, 0)
	for _, calc := range a.sysAddRateAttr {
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

func (a *AttrSystemAdapter) calcTotalSysAddRate(ctx context.Context) {
	if len(a.addRateCalcs) == 0 {
		for _, calc := range a.sysAddRateAttr {
			if calc != nil {
				calc.Reset()
			}
		}
		return
	}
	totalCalc := icalc.NewFightAttrCalc()
	for _, calc := range a.sysAttr {
		if calc == nil {
			continue
		}
		totalCalc.AddCalc(calc)
	}
	for sysID, calculator := range a.addRateCalcs {
		if calculator == nil {
			continue
		}
		targetCalc := a.ensureSysAddRateCalc(sysID)
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

func (a *AttrSystemAdapter) pushAttrDataToClient(ctx context.Context, playerRole iface.IPlayerRole, syncData *protocol.SyncAttrData) {
	if playerRole == nil || syncData == nil {
		return
	}
	cfg := getAttrPushConfig()
	filtered := a.filterSyncAttrData(syncData, cfg)
	if filtered == nil {
		return
	}
	resp := &protocol.S2CAttrDataReq{
		AttrData: filtered,
	}
	if cfg.IncludeSysPowerMap {
		resp.SysPowerMap = a.snapshotSysPowerMap()
	}
	sessionID := playerRole.GetSessionId()
	if err := a.networkGateway.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CAttrData), resp); err != nil {
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

func (a *AttrSystemAdapter) filterSyncAttrData(data *protocol.SyncAttrData, cfg *jsonconf.AttrPushConfig) *protocol.SyncAttrData {
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

func (a *AttrSystemAdapter) snapshotSysPowerMap() map[uint32]int64 {
	if len(a.sysPowerMap) == 0 {
		return nil
	}
	result := make(map[uint32]int64, len(a.sysPowerMap))
	for k, v := range a.sysPowerMap {
		result[k] = v
	}
	return result
}
