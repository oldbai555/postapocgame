/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc: 属性系统实现
**/

package entitysystem

import (
	"context"

	"postapocgame/server/internal/attrcalc"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/attrpower"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	dungeonattrcalc "postapocgame/server/service/dungeonserver/internel/entitysystem/attrcalc"
	"postapocgame/server/service/dungeonserver/internel/gameserverlink"
	"postapocgame/server/service/dungeonserver/internel/iface"

	"google.golang.org/protobuf/proto"
)

var _ iface.IAttrSys = (*AttrSys)(nil)

// AttrSys 属性系统
type AttrSys struct {
	owner       iface.IEntity
	attrSet     *attrcalc.AttrSet
	fightAttr   *attrcalc.FightAttrCalc
	extraAttr   *attrcalc.ExtraAttrCalc
	addRateAttr *attrcalc.FightAttrCalc

	dirty           bool
	lastMaxHP       attrdef.AttrValue
	extraUpdateMask map[attrdef.AttrType]bool // 跟踪非战斗属性的变化
	bInitFinish     bool                      // 初始化完成标志（初始化完成前不广播属性）
}

// NewAttrSys 创建属性系统
func NewAttrSys(owner iface.IEntity) *AttrSys {
	return &AttrSys{
		owner:           owner,
		attrSet:         attrcalc.NewAttrSet(),
		fightAttr:       attrcalc.NewFightAttrCalc(),
		extraAttr:       attrcalc.NewExtraAttrCalc(),
		addRateAttr:     attrcalc.NewFightAttrCalc(),
		dirty:           true,
		extraUpdateMask: make(map[attrdef.AttrType]bool),
		bInitFinish:     false,
	}
}

// ApplySyncData 应用来自逻辑服的系统属性数据
func (as *AttrSys) ApplySyncData(syncData *protocol.SyncAttrData) {
	if syncData == nil {
		return
	}
	if syncData.AttrData != nil {
		for sysID, attrVec := range syncData.AttrData {
			calc := as.attrSet.GetIncAttr(sysID, true, true)
			if attrVec == nil {
				continue
			}
			for _, attr := range attrVec.Attrs {
				if attr == nil {
					continue
				}
				attrType := attrdef.AttrType(attr.Type)
				calc.AddValue(attrType, attrdef.AttrValue(attr.Value))
			}
		}
	}
	if vec := syncData.AddRateAttr; vec != nil {
		as.addRateAttr.Reset()
		for _, attr := range vec.Attrs {
			if attr == nil {
				continue
			}
			as.addRateAttr.AddValue(attrdef.AttrType(attr.Type), attrdef.AttrValue(attr.Value))
		}
	}
	as.MarkDirty()
}

// GetAttrValue 获取属性值
func (as *AttrSys) GetAttrValue(attrType attrdef.AttrType) attrdef.AttrValue {
	as.ensureAggregated()
	if attrdef.IsCombatAttr(attrType) {
		return as.fightAttr.GetValue(attrType)
	}
	if attrdef.IsExtraAttr(attrType) {
		return as.extraAttr.GetValue(attrType)
	}
	return 0
}

func (as *AttrSys) SetAttrValue(attrType attrdef.AttrType, value attrdef.AttrValue) {
	as.ensureAggregated()
	if attrdef.IsCombatAttr(attrType) {
		as.fightAttr.SetValue(attrType, value)
		return
	}
	if attrdef.IsExtraAttr(attrType) {
		as.extraAttr.SetValue(attrType, value)
		// 标记非战斗属性变化
		as.extraUpdateMask[attrType] = true
	}
}

func (as *AttrSys) AddAttrValue(attrType attrdef.AttrType, delta attrdef.AttrValue) {
	as.ensureAggregated()
	if attrdef.IsCombatAttr(attrType) {
		as.fightAttr.AddValue(attrType, delta)
		return
	}
	if attrdef.IsExtraAttr(attrType) {
		as.extraAttr.AddValue(attrType, delta)
	}
}

// GetAllCombatAttrs 获取所有战斗属性
func (as *AttrSys) GetAllCombatAttrs() map[attrdef.AttrType]attrdef.AttrValue {
	as.ensureAggregated()
	result := make(map[attrdef.AttrType]attrdef.AttrValue)
	as.fightAttr.DoRange(func(attrType attrdef.AttrType, value attrdef.AttrValue) {
		result[attrType] = value
	})
	return result
}

// GetAllExtraAttrs 获取所有非战斗属性
func (as *AttrSys) GetAllExtraAttrs() map[attrdef.AttrType]attrdef.AttrValue {
	as.ensureAggregated()
	return as.extraAttr.GetAll()
}

// ResetCombatAttrs 重置战斗属性
func (as *AttrSys) ResetCombatAttrs() {
	as.fightAttr.Reset()
}

// ResetExtraAttrs 重置非战斗属性
func (as *AttrSys) ResetExtraAttrs() {
	as.extraAttr.Reset()
}

// ResetAll 重置所有属性
func (as *AttrSys) ResetAll() {
	as.fightAttr.Reset()
	as.extraAttr.Reset()
}

// BatchSetAttrs 批量设置属性
func (as *AttrSys) BatchSetAttrs(attrs map[attrdef.AttrType]attrdef.AttrValue) {
	as.ensureAggregated()
	for attrType, value := range attrs {
		if attrdef.IsCombatAttr(attrType) {
			as.fightAttr.SetValue(attrType, value)
		} else if attrdef.IsExtraAttr(attrType) {
			as.extraAttr.SetValue(attrType, value)
		}
	}
}

// BatchAddAttrs 批量增加属性
func (as *AttrSys) BatchAddAttrs(attrs map[attrdef.AttrType]attrdef.AttrValue) {
	as.ensureAggregated()
	for attrType, delta := range attrs {
		if attrdef.IsCombatAttr(attrType) {
			as.fightAttr.AddValue(attrType, delta)
		} else if attrdef.IsExtraAttr(attrType) {
			as.extraAttr.AddValue(attrType, delta)
		}
	}
}

// GetHP 获取当前生命值
func (as *AttrSys) GetHP() int64 {
	as.ensureAggregated()
	val := as.GetAttrValue(attrdef.AttrHP)
	return int64(val)
}

// GetMaxHP 获取最大生命值
func (as *AttrSys) GetMaxHP() int64 {
	as.ensureAggregated()
	val := as.GetAttrValue(attrdef.AttrMaxHP)
	return int64(val)
}

// GetMP 获取当前魔法值
func (as *AttrSys) GetMP() int64 {
	as.ensureAggregated()
	val := as.GetAttrValue(attrdef.AttrMP)
	return int64(val)
}

// GetMaxMP 获取最大魔法值
func (as *AttrSys) GetMaxMP() int64 {
	as.ensureAggregated()
	val := as.GetAttrValue(attrdef.AttrMaxMP)
	return int64(val)
}

// GetAttack 获取攻击力
func (as *AttrSys) GetAttack() int64 {
	as.ensureAggregated()
	val := as.GetAttrValue(attrdef.AttrAttack)
	return int64(val)
}

// GetDefense 获取防御力
func (as *AttrSys) GetDefense() int64 {
	as.ensureAggregated()
	val := as.GetAttrValue(attrdef.AttrDefense)
	return int64(val)
}

// GetSpeed 获取速度
func (as *AttrSys) GetSpeed() int64 {
	as.ensureAggregated()
	val := as.GetAttrValue(attrdef.AttrSpeed)
	return int64(val)
}

// GetCritRate 获取暴击率（万分比）
func (as *AttrSys) GetCritRate() int64 {
	as.ensureAggregated()
	val := as.GetAttrValue(attrdef.AttrCritRate)
	return int64(val)
}

// GetDodgeRate 获取闪避率（万分比）
func (as *AttrSys) GetDodgeRate() int64 {
	as.ensureAggregated()
	val := as.GetAttrValue(attrdef.AttrDodgeRate)
	return int64(val)
}

// SetHP 设置当前生命值
func (as *AttrSys) SetHP(hp int64) {
	as.ensureAggregated()
	// 限制HP不超过MaxHP
	maxHP := as.GetMaxHP()
	if hp > maxHP {
		hp = maxHP
	}
	if hp < 0 {
		hp = 0
	}
	as.SetAttrValue(attrdef.AttrHP, hp)
}

// SetMP 设置当前魔法值
func (as *AttrSys) SetMP(mp int64) {
	as.ensureAggregated()
	// 限制MP不超过MaxMP
	maxMP := as.GetMaxMP()
	if mp > maxMP {
		mp = maxMP
	}
	if mp < 0 {
		mp = 0
	}
	as.SetAttrValue(attrdef.AttrMP, attrdef.AttrValue(mp))
}

// AddHP 增加生命值
func (as *AttrSys) AddHP(delta int64) {
	as.ensureAggregated()
	currentHP := as.GetHP()
	as.SetHP(currentHP + delta)
}

// AddMP 增加魔法值
func (as *AttrSys) AddMP(delta int64) {
	as.ensureAggregated()
	currentMP := as.GetMP()
	as.SetMP(currentMP + delta)
}

func (as *AttrSys) ensureAggregated() {
	if !as.dirty {
		return
	}
	as.ResetProperty()
}

// ResetProperty 重新计算战斗属性并广播
func (as *AttrSys) ResetProperty() {
	if !as.dirty {
		return
	}
	as.dirty = false
	temp := attrcalc.NewFightAttrCalc()
	as.attrSet.ResetProperty(temp)
	temp.AddCalc(as.addRateAttr)
	attrcalc.ApplyConversions(temp)
	attrcalc.ApplyPercentages(temp)
	as.applyFightAttr(temp)
	as.broadcastAttrData()
}

// MarkDirty 标记需要重新计算
func (as *AttrSys) MarkDirty() {
	as.dirty = true
}

// ResetSysAttr 重置指定系统的属性（通过注册管理器计算）
func (as *AttrSys) ResetSysAttr(sysId uint32) {
	// 1. 获取增量计算回调
	if incFn := dungeonattrcalc.GetIncAttrCalcFn(sysId); incFn != nil {
		calc := as.attrSet.GetIncAttr(sysId, true, true)
		incFn(as.owner, calc)
		as.MarkDirty()
	}

	// 2. 获取减量计算回调
	if decFn := dungeonattrcalc.GetDecAttrCalcFn(sysId); decFn != nil {
		calc := as.attrSet.GetDecAttr(sysId, true, true)
		decFn(as.owner, calc)
		as.MarkDirty()
	}
}

func (as *AttrSys) applyFightAttr(calc *attrcalc.FightAttrCalc) {
	if calc == nil {
		return
	}
	oldMax := as.lastMaxHP
	newMax := calc.GetValue(attrdef.AttrMaxHP)
	as.fightAttr.Copy(calc)
	as.lastMaxHP = newMax
	as.adjustHP(oldMax, newMax)
}

func (as *AttrSys) adjustHP(oldMax, newMax attrdef.AttrValue) {
	currentHP := as.extraAttr.GetValue(attrdef.AttrHP)
	if newMax <= 0 {
		as.extraAttr.SetValue(attrdef.AttrHP, 0)
		return
	}
	if currentHP == 0 {
		as.extraAttr.SetValue(attrdef.AttrHP, newMax)
		return
	}
	if oldMax <= 0 {
		as.extraAttr.SetValue(attrdef.AttrHP, newMax)
		return
	}
	rate := float64(currentHP) / float64(oldMax)
	reBalanced := attrdef.AttrValue(int64(rate * float64(newMax)))
	if reBalanced > newMax {
		reBalanced = newMax
	}
	if reBalanced < 0 {
		reBalanced = 0
	}
	as.extraAttr.SetValue(attrdef.AttrHP, reBalanced)
}

func (as *AttrSys) broadcastAttrData() {
	// 初始化完成前不广播属性
	if !as.bInitFinish {
		return
	}
	syncData := as.buildSyncAttrData()
	if syncData == nil {
		return
	}
	resp := &protocol.S2CAttrDataReq{
		AttrData:    syncData,
		SysPowerMap: map[uint32]int64{uint32(protocol.SaAttrSys_SaAttrSysNil): as.calcPower()},
	}
	_ = as.owner.SendProtoMessage(uint16(protocol.S2CProtocol_S2CAttrData), resp)
	as.syncBackToGameServer(syncData)
}

func (as *AttrSys) buildSyncAttrData() *protocol.SyncAttrData {
	attrMap := as.attrSet.PackToMap()
	addRateVec := &protocol.AttrVec{}
	as.addRateAttr.DoRange(func(attrType attrdef.AttrType, value attrdef.AttrValue) {
		addRateVec.Attrs = append(addRateVec.Attrs, &protocol.AttrSt{
			Type:  uint32(attrType),
			Value: int64(value),
		})
	})
	if len(attrMap) == 0 && len(addRateVec.Attrs) == 0 {
		return nil
	}
	syncData := &protocol.SyncAttrData{
		AttrData: make(map[uint32]*protocol.AttrVec, len(attrMap)),
	}
	for sysID, attrs := range attrMap {
		vec := &protocol.AttrVec{Attrs: make([]*protocol.AttrSt, 0, len(attrs))}
		for attrType, value := range attrs {
			vec.Attrs = append(vec.Attrs, &protocol.AttrSt{
				Type:  uint32(attrType),
				Value: int64(value),
			})
		}
		if len(vec.Attrs) > 0 {
			syncData.AttrData[sysID] = vec
		}
	}
	if len(addRateVec.Attrs) > 0 {
		syncData.AddRateAttr = addRateVec
	}
	return syncData
}

func (as *AttrSys) calcPower() int64 {
	return attrpower.CalcPower(as.fightAttr, as.getOwnerJob())
}

func (as *AttrSys) getOwnerJob() uint32 {
	if ownerWithJob, ok := as.owner.(interface {
		GetJobId() uint32
	}); ok {
		return ownerWithJob.GetJobId()
	}
	return 0
}

func (as *AttrSys) syncBackToGameServer(syncData *protocol.SyncAttrData) {
	role, ok := as.owner.(iface.IRole)
	if !ok || syncData == nil {
		return
	}
	req := &protocol.D2GSyncAttrsReq{
		SessionId: role.GetSessionId(),
		RoleId:    as.owner.GetId(),
		SyncData:  syncData,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		log.Errorf("marshal D2GSyncAttrsReq failed: %v", err)
		return
	}
	if err := gameserverlink.CallGameServer(context.Background(), role.GetSessionId(), uint16(protocol.D2GRpcProtocol_D2GSyncAttrs), data); err != nil {
		log.Errorf("sync attrs back to GameServer failed: %v", err)
	}
}

// SetInitFinish 标记初始化完成（初始化完成前不广播属性）
func (as *AttrSys) SetInitFinish() {
	as.bInitFinish = true
}

// RunOne 每帧更新（由实体 RunOne 调用）
func (as *AttrSys) RunOne() {
	as.ResetProperty()
	as.CheckAndSyncProp()
}

// CheckAndSyncProp 检查并同步非战斗属性的变化
func (as *AttrSys) CheckAndSyncProp() {
	if len(as.extraUpdateMask) == 0 {
		return
	}
	// 初始化完成前不广播
	if !as.bInitFinish {
		// 清空 mask，但不广播
		as.extraUpdateMask = make(map[attrdef.AttrType]bool)
		return
	}

	// 检查是否为角色实体（只有角色需要广播非战斗属性）
	_, isRole := as.owner.(iface.IRole)
	if !isRole {
		// 非角色实体不需要广播非战斗属性
		as.extraUpdateMask = make(map[attrdef.AttrType]bool)
		return
	}

	// 收集需要广播的非战斗属性
	// 注意：当前实现简化处理，直接通过 S2CAttrData 协议推送所有属性
	// 如果需要区分广播和自看属性，可以参考 jjyz 的实现使用 ExtraAttrDesc 配置
	// 这里先简化实现，后续可以根据需要扩展

	// 清空 mask
	as.extraUpdateMask = make(map[attrdef.AttrType]bool)
}
