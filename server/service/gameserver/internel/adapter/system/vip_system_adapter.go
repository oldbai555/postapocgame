package system

import (
	"context"
	"postapocgame/server/service/gameserver/internel/core/iface"

	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	vipdomain "postapocgame/server/service/gameserver/internel/domain/vip"
)

// VipSystemAdapter VIP 系统适配器（负责初始化 VipData 与提供特权查询）
type VipSystemAdapter struct {
	*BaseSystemAdapter
}

// NewVipSystemAdapter 创建 VIP 系统适配器
func NewVipSystemAdapter() *VipSystemAdapter {
	return &VipSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysVip)),
	}
}

// OnInit 初始化 VIP 数据
func (a *VipSystemAdapter) OnInit(ctx context.Context) {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("vip sys OnInit get role err:%v", err)
		return
	}
	vipdomain.EnsureVipData(playerRole.GetBinaryData())
}

// GetVipData 获取 VIP 数据
func (a *VipSystemAdapter) GetVipData(ctx context.Context) *protocol.SiVipData {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		return nil
	}
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		return nil
	}
	return vipdomain.EnsureVipData(binaryData)
}

// GetLevel 获取当前 VIP 等级
func (a *VipSystemAdapter) GetLevel(ctx context.Context) uint32 {
	data := a.GetVipData(ctx)
	if data == nil {
		return 0
	}
	return data.Level
}

// GetExp 获取当前 VIP 经验
func (a *VipSystemAdapter) GetExp(ctx context.Context) uint32 {
	data := a.GetVipData(ctx)
	if data == nil {
		return 0
	}
	return data.Exp
}

// GetPrivilegeValue 获取指定特权的值
func (a *VipSystemAdapter) GetPrivilegeValue(ctx context.Context, privilegeType uint32) uint32 {
	data := a.GetVipData(ctx)
	if data == nil {
		return 0
	}
	vipConfig, ok := jsonconf.GetConfigManager().GetVipConfig(data.Level)
	if !ok || vipConfig == nil {
		return 0
	}
	for _, p := range vipConfig.Privileges {
		if p.Type == privilegeType {
			return p.Value
		}
	}
	return 0
}

// HasPrivilege 检查是否有指定特权
func (a *VipSystemAdapter) HasPrivilege(ctx context.Context, privilegeType uint32) bool {
	return a.GetPrivilegeValue(ctx, privilegeType) > 0
}

// 确保实现 ISystem 接口
var _ iface.ISystem = (*VipSystemAdapter)(nil)
