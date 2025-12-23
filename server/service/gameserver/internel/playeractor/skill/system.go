package skill

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/playeractor/sysbase"
)

type SystemAdapter struct {
	*sysbase.BaseSystem
	rt *deps.Runtime
}

// NewSkillSystemAdapter 创建技能系统适配器
func NewSkillSystemAdapter(rt *deps.Runtime) *SystemAdapter {
	return &SystemAdapter{
		BaseSystem: sysbase.NewBaseSystem(uint32(protocol.SystemId_SysSkill)),
		rt:         rt,
	}
}

// OnInit 系统初始化
func (a *SystemAdapter) OnInit(ctx context.Context) {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("skill sys OnInit get role err:%v", err)
		return
	}
	initSkillDataUC := NewInitSkillDataUseCase(a.rt)
	if err := initSkillDataUC.Execute(ctx, playerRole.GetPlayerRoleId(), playerRole.GetJob()); err != nil {
		log.Errorf("skill sys OnInit init skill data err:%v", err)
		return
	}
}

func (a *SystemAdapter) GetSkillMap(ctx context.Context) (map[uint32]uint32, error) {
	skillData, err := a.rt.PlayerRepo().GetSkillData(ctx)
	if err != nil {
		return nil, err
	}
	return skillData.SkillMap, nil
}

// EnsureISystem 确保 SkillSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*SystemAdapter)(nil)

// GetSkillSys 获取技能系统
func GetSkillSys(ctx context.Context) *SystemAdapter {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysSkill))
	if system == nil {
		log.Errorf("not found system [%v]", protocol.SystemId_SysSkill)
		return nil
	}
	sys, ok := system.(*SystemAdapter)
	if !ok {
		log.Errorf("invalid system type for [%v]", protocol.SystemId_SysSkill)
		return nil
	}
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error", protocol.SystemId_SysSkill)
		return nil
	}
	return sys
}
func RegisterSystemFactory(rt *deps.Runtime) {
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysSkill), func() iface.ISystem {
		return NewSkillSystemAdapter(rt)
	})
}
