package entitysystem

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
)

// VipSys VIP系统
type VipSys struct {
	*BaseSystem
	vipData *protocol.SiVipData
}

// NewVipSys 创建VIP系统
func NewVipSys() *VipSys {
	return &VipSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysVip)),
	}
}

func GetVipSys(ctx context.Context) *VipSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysVip))
	if system == nil {
		return nil
	}
	vipSys, ok := system.(*VipSys)
	if !ok || !vipSys.IsOpened() {
		return nil
	}
	return vipSys
}

// OnInit 系统初始化
func (vs *VipSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("vip sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果vip_data不存在，则初始化
	if binaryData.VipData == nil {
		binaryData.VipData = &protocol.SiVipData{
			Level: 0,
			Exp:   0,
		}
	}
	vs.vipData = binaryData.VipData
	// uint32类型不可能小于0，无需检查
}

// GetVipData 获取VIP数据
func (vs *VipSys) GetVipData() *protocol.SiVipData {
	return vs.vipData
}

// GetLevel 获取当前VIP等级
func (vs *VipSys) GetLevel() uint32 {
	if vs.vipData == nil {
		return 0
	}
	return vs.vipData.Level
}

// GetExp 获取当前VIP经验
func (vs *VipSys) GetExp() uint32 {
	if vs.vipData == nil {
		return 0
	}
	return vs.vipData.Exp
}

// AddVipExp 添加VIP经验值
func (vs *VipSys) AddVipExp(ctx context.Context, exp uint32) error {
	if exp == 0 {
		return nil
	}

	if vs.vipData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "vip data not initialized")
	}

	// 添加经验
	vs.vipData.Exp += exp

	// 检查是否升级
	if err := vs.CheckVipLevelUp(ctx); err != nil {
		return err
	}

	return nil
}

// CheckVipLevelUp 检查并处理VIP升级逻辑
func (vs *VipSys) CheckVipLevelUp(ctx context.Context) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	if vs.vipData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "vip data not initialized")
	}

	// 循环检查升级（可能一次获得大量经验，连续升级）
	for {
		// 获取当前等级的配置
		vipConfig, ok := jsonconf.GetConfigManager().GetVipConfig(vs.vipData.Level)
		if !ok {
			// 没有更高等级的配置，已达到最高等级
			break
		}

		// 检查是否满足升级条件
		if uint64(vs.vipData.Exp) < vipConfig.ExpNeeded {
			break
		}

		// 扣除升级所需经验
		vs.vipData.Exp -= uint32(vipConfig.ExpNeeded)

		// 升级
		vs.vipData.Level++

		log.Infof("Player VIP level up: PlayerID=%d, NewVipLevel=%d, RemainingExp=%d",
			playerRole.GetPlayerRoleId(), vs.vipData.Level, vs.vipData.Exp)

		// VIP升级不发放奖励，但可以发布事件供其他系统订阅
		// 例如：背包扩容、副本次数增加等特权功能
	}

	return nil
}

// GetPrivilegeValue 获取指定特权的值
func (vs *VipSys) GetPrivilegeValue(privilegeType uint32) uint32 {
	if vs.vipData == nil {
		return 0
	}

	// 获取当前VIP等级的配置
	vipConfig, ok := jsonconf.GetConfigManager().GetVipConfig(vs.vipData.Level)
	if !ok {
		return 0
	}

	// 查找指定类型的特权
	for _, privilege := range vipConfig.Privileges {
		if privilege.Type == privilegeType {
			return privilege.Value
		}
	}

	return 0
}

// HasPrivilege 检查是否有指定特权
func (vs *VipSys) HasPrivilege(privilegeType uint32) bool {
	return vs.GetPrivilegeValue(privilegeType) > 0
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysVip), func() iface.ISystem {
		return NewVipSys()
	})
}
