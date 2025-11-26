package entitysystem

import (
	"context"

	"postapocgame/server/internal"
	icalc "postapocgame/server/internal/attrcalc"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/attrpower"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/iface"
	gameattrcalc "postapocgame/server/service/gameserver/internel/playeractor/entitysystem/attrcalc"
)

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

// AttrSys 属性系统汇总
type AttrSys struct {
	*BaseSystem
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
}

// NewAttrSys 创建属性系统
func NewAttrSys() *AttrSys {
	return &AttrSys{
		BaseSystem:     NewBaseSystem(uint32(protocol.SystemId_SysAttr)),
		attrDataMap:    make(map[uint32]*protocol.AttrVec),
		dirtySystems:   make(map[uint32]bool),
		calculators:    make(map[uint32]gameattrcalc.Calculator),
		addRateCalcs:   make(map[uint32]gameattrcalc.AddRateCalculator),
		sysAttr:        make(map[uint32]*icalc.FightAttrCalc),
		sysAddRateAttr: make(map[uint32]*icalc.FightAttrCalc),
		extraAttr:      icalc.NewExtraAttrCalc(),
		fightAttr:      icalc.NewFightAttrCalc(),
		extraDirty:     make(map[attrdef.AttrType]struct{}),
		sysPowerMap:    make(map[uint32]int64),
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
	as.cloneCalculators(ctx)
	as.cloneAddRateCalculators(ctx)
}

func (as *AttrSys) cloneCalculators(ctx context.Context) {
	calculators := gameattrcalc.CloneCalculators(ctx)
	if len(calculators) == 0 {
		as.calculators = make(map[uint32]gameattrcalc.Calculator)
		return
	}
	as.calculators = calculators
}

func (as *AttrSys) cloneAddRateCalculators(ctx context.Context) {
	calculators := gameattrcalc.CloneAddRateCalculators(ctx)
	if len(calculators) == 0 {
		as.addRateCalcs = make(map[uint32]gameattrcalc.AddRateCalculator)
		return
	}
	as.addRateCalcs = calculators
}

// MarkDirty 标记需要重算的系统
func (as *AttrSys) MarkDirty(saAttrSysId uint32) {
	as.dirtySystems[saAttrSysId] = true
	as.sysAttrChanged = true
}

// ResetSysAttr 对外暴露的系统属性重算入口
func (as *AttrSys) ResetSysAttr(saAttrSysId uint32) {
	as.MarkDirty(saAttrSysId)
}

// CalculateAllAttrs 计算所有系统的属性
func (as *AttrSys) CalculateAllAttrs(ctx context.Context) map[uint32]*protocol.AttrVec {
	result := make(map[uint32]*protocol.AttrVec)
	for sysID := range as.calculators {
		if vec, exists := as.attrDataMap[sysID]; exists && !as.isSystemDirty(sysID) {
			result[sysID] = vec
			continue
		}
		attrVec, changed := as.calculateSystemAttr(ctx, sysID)
		if attrVec == nil {
			continue
		}
		result[sysID] = attrVec
		if changed {
			as.sysAttrChanged = true
		}
		as.clearSystemDirty(sysID)
	}
	as.dirtySystems = make(map[uint32]bool)
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

	changedAttrs := make(map[uint32]*protocol.AttrVec)
	for saAttrSysId := range as.dirtySystems {
		attrVec, changed := as.calculateSystemAttr(ctx, saAttrSysId)
		as.clearSystemDirty(saAttrSysId)
		if attrVec == nil || !changed {
			continue
		}
		changedAttrs[saAttrSysId] = attrVec
	}
	as.dirtySystems = make(map[uint32]bool)
	if len(changedAttrs) == 0 {
		return
	}
	as.sysAttrChanged = true
	as.calcSysPowerMap(ctx)
	as.syncAttrsToDungeonServer(ctx, playerRole, changedAttrs)
}

// syncAttrsToDungeonServer 同步属性到DungeonServer
func (as *AttrSys) syncAttrsToDungeonServer(ctx context.Context, playerRole iface.IPlayerRole, changedAttrs map[uint32]*protocol.AttrVec) {
	as.calcTotalSysAddRate(ctx)
	// 构建同步数据
	syncData := &protocol.SyncAttrData{
		AttrData: make(map[uint32]*protocol.AttrVec),
	}
	for saAttrSysId, attrVec := range changedAttrs {
		syncData.AttrData[saAttrSysId] = attrVec
	}
	if addRateVec := as.buildAddRateAttrVec(); addRateVec != nil {
		syncData.AddRateAttr = addRateVec
	}
	as.pushAttrDataToClient(playerRole, syncData)

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
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
	})
}

func (as *AttrSys) calculateSystemAttr(ctx context.Context, sysID uint32) (*protocol.AttrVec, bool) {
	calculator, exists := as.calculators[sysID]
	if !exists || calculator == nil {
		_, had := as.attrDataMap[sysID]
		delete(as.attrDataMap, sysID)
		delete(as.sysAttr, sysID)
		return nil, had
	}
	rawAttrs := calculator.CalculateAttrs(ctx)
	if len(rawAttrs) == 0 {
		delete(as.attrDataMap, sysID)
		delete(as.sysAttr, sysID)
		return nil, true
	}
	attrVec := &protocol.AttrVec{
		Attrs: cloneAttrList(rawAttrs),
	}
	changed := !attrVecEquals(as.attrDataMap[sysID], attrVec)
	as.attrDataMap[sysID] = attrVec
	as.refreshSysCalc(sysID, attrVec)
	return attrVec, changed
}

func (as *AttrSys) refreshSysCalc(sysID uint32, attrVec *protocol.AttrVec) {
	if attrVec == nil {
		return
	}
	calc := as.ensureSysCalc(sysID)
	calc.Reset()
	for _, attr := range attrVec.Attrs {
		if attr == nil {
			continue
		}
		attrType := attrdef.AttrType(attr.Type)
		calc.AddValue(attrType, attrdef.AttrValue(attr.Value))
	}
}

func (as *AttrSys) ensureSysCalc(sysID uint32) *icalc.FightAttrCalc {
	if calc, ok := as.sysAttr[sysID]; ok && calc != nil {
		return calc
	}
	calc := icalc.NewFightAttrCalc()
	as.sysAttr[sysID] = calc
	return calc
}

func (as *AttrSys) ensureSysAddRateCalc(sysID uint32) *icalc.FightAttrCalc {
	if calc, ok := as.sysAddRateAttr[sysID]; ok && calc != nil {
		return calc
	}
	calc := icalc.NewFightAttrCalc()
	as.sysAddRateAttr[sysID] = calc
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

func (as *AttrSys) calcSysPowerMap(ctx context.Context) {
	if len(as.sysAttr) == 0 {
		as.sysPowerMap = make(map[uint32]int64)
		return
	}
	var job uint32
	if playerRole, err := GetIPlayerRoleByContext(ctx); err == nil && playerRole != nil {
		job = playerRole.GetJob()
	}
	sysPower := make(map[uint32]int64, len(as.sysAttr))
	for sysID, calc := range as.sysAttr {
		if calc == nil {
			continue
		}
		temp := icalc.NewFightAttrCalc()
		temp.AddCalc(calc)
		if addRateCalc := as.sysAddRateAttr[sysID]; addRateCalc != nil {
			temp.AddCalc(addRateCalc)
		}
		icalc.ApplyConversions(temp)
		icalc.ApplyPercentages(temp)
		sysPower[sysID] = attrpower.CalcPower(temp, job)
	}
	as.sysPowerMap = sysPower
}

func (as *AttrSys) isSystemDirty(sysID uint32) bool {
	_, ok := as.dirtySystems[sysID]
	return ok
}

func (as *AttrSys) clearSystemDirty(sysID uint32) {
	delete(as.dirtySystems, sysID)
}

func (as *AttrSys) buildAddRateAttrVec() *protocol.AttrVec {
	if len(as.sysAddRateAttr) == 0 {
		return nil
	}
	attrs := make([]*protocol.AttrSt, 0)
	for _, calc := range as.sysAddRateAttr {
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

func (as *AttrSys) calcTotalSysAddRate(ctx context.Context) {
	if len(as.addRateCalcs) == 0 {
		for _, calc := range as.sysAddRateAttr {
			if calc != nil {
				calc.Reset()
			}
		}
		return
	}
	totalCalc := icalc.NewFightAttrCalc()
	for _, calc := range as.sysAttr {
		if calc == nil {
			continue
		}
		totalCalc.AddCalc(calc)
	}
	for sysID, calculator := range as.addRateCalcs {
		if calculator == nil {
			continue
		}
		targetCalc := as.ensureSysAddRateCalc(sysID)
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

func (as *AttrSys) pushAttrDataToClient(playerRole iface.IPlayerRole, syncData *protocol.SyncAttrData) {
	if playerRole == nil || syncData == nil {
		return
	}
	cfg := getAttrPushConfig()
	filtered := as.filterSyncAttrData(syncData, cfg)
	if filtered == nil {
		return
	}
	resp := &protocol.S2CAttrDataReq{
		AttrData: filtered,
	}
	if cfg.IncludeSysPowerMap {
		resp.SysPowerMap = as.snapshotSysPowerMap()
	}
	if err := playerRole.SendProtoMessage(uint16(protocol.S2CProtocol_S2CAttrData), resp); err != nil {
		log.Errorf("push attr data to client failed: %v", err)
	}
}

func (as *AttrSys) filterSyncAttrData(data *protocol.SyncAttrData, cfg *jsonconf.AttrPushConfig) *protocol.SyncAttrData {
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

func (as *AttrSys) snapshotSysPowerMap() map[uint32]int64 {
	if len(as.sysPowerMap) == 0 {
		return nil
	}
	result := make(map[uint32]int64, len(as.sysPowerMap))
	for k, v := range as.sysPowerMap {
		result[k] = v
	}
	return result
}

// PushFullAttrData 将当前缓存的属性完整推送给客户端
func (as *AttrSys) PushFullAttrData(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("push full attr data failed: %v", err)
		return
	}
	if len(as.attrDataMap) == 0 {
		return
	}
	as.calcSysPowerMap(ctx)
	syncData := &protocol.SyncAttrData{
		AttrData: cloneAttrDataMap(as.attrDataMap),
	}
	if len(syncData.AttrData) == 0 {
		return
	}
	as.pushAttrDataToClient(playerRole, syncData)
}

// PushSyncDataToClient 按需推送指定的属性数据
func (as *AttrSys) PushSyncDataToClient(ctx context.Context, syncData *protocol.SyncAttrData) {
	if syncData == nil {
		return
	}
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("push attr data failed: %v", err)
		return
	}
	as.calcSysPowerMap(ctx)
	as.pushAttrDataToClient(playerRole, syncData)
}

// ApplyDungeonSyncData 同步战斗服回写的属性结果
func (as *AttrSys) ApplyDungeonSyncData(syncData *protocol.SyncAttrData) {
	if syncData == nil || as.fightAttr == nil {
		return
	}
	as.fightAttr.Reset()
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
			as.fightAttr.AddValue(attrType, attrdef.AttrValue(attr.Value))
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
			as.fightAttr.AddValue(attrType, attrdef.AttrValue(attr.Value))
		}
	}
}

func cloneAttrDataMap(src map[uint32]*protocol.AttrVec) map[uint32]*protocol.AttrVec {
	if len(src) == 0 {
		return nil
	}
	result := make(map[uint32]*protocol.AttrVec, len(src))
	for sysID, vec := range src {
		if vec == nil {
			continue
		}
		result[sysID] = &protocol.AttrVec{
			Attrs: cloneAttrList(vec.Attrs),
		}
	}
	return result
}
