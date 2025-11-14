package entitysystem

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
	"time"
)

// NewbieSys 新手保护系统
type NewbieSys struct {
	*BaseSystem
	newbieData *protocol.SiNewbieData
}

// NewNewbieSys 创建新手保护系统
func NewNewbieSys() *NewbieSys {
	return &NewbieSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysNewbie)),
	}
}

func GetNewbieSys(ctx context.Context) *NewbieSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysNewbie))
	if system == nil {
		return nil
	}
	newbieSys, ok := system.(*NewbieSys)
	if !ok || !newbieSys.IsOpened() {
		return nil
	}
	return newbieSys
}

// OnInit 系统初始化
func (ns *NewbieSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("newbie sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果newbie_data不存在，则初始化
	if binaryData.NewbieData == nil {
		binaryData.NewbieData = &protocol.SiNewbieData{
			IsNewbie:      true,
			NewbieEndTime: time.Now().Add(7 * 24 * time.Hour).Unix(), // 默认7天新手保护期
			GuideProgress: make(map[uint32]bool),
		}
	}
	ns.newbieData = binaryData.NewbieData

	// 如果GuideProgress为空，初始化为空map
	if ns.newbieData.GuideProgress == nil {
		ns.newbieData.GuideProgress = make(map[uint32]bool)
	}

	// 检查新手保护期是否已过期
	ns.checkNewbieStatus()

	log.Infof("NewbieSys initialized: IsNewbie=%v, NewbieEndTime=%d", ns.newbieData.IsNewbie, ns.newbieData.NewbieEndTime)
}

// checkNewbieStatus 检查新手保护期状态
func (ns *NewbieSys) checkNewbieStatus() {
	now := time.Now().Unix()
	if ns.newbieData.NewbieEndTime > 0 && now >= ns.newbieData.NewbieEndTime {
		ns.newbieData.IsNewbie = false
	}
}

// IsNewbie 检查是否处于新手保护期
func (ns *NewbieSys) IsNewbie() bool {
	ns.checkNewbieStatus()
	return ns.newbieData.IsNewbie
}

// IsNewbieByLevel 根据等级检查是否处于新手保护期（默认10级以下为新手）
func (ns *NewbieSys) IsNewbieByLevel(level uint32) bool {
	// 如果等级小于10，强制为新手
	if level < 10 {
		ns.newbieData.IsNewbie = true
		return true
	}
	return ns.IsNewbie()
}

// SetNewbieEndTime 设置新手保护期结束时间
func (ns *NewbieSys) SetNewbieEndTime(endTime int64) {
	ns.newbieData.NewbieEndTime = endTime
	ns.newbieData.IsNewbie = true
}

// EndNewbieProtection 结束新手保护期
func (ns *NewbieSys) EndNewbieProtection() {
	ns.newbieData.IsNewbie = false
	ns.newbieData.NewbieEndTime = 0
}

// UpdateGuideProgress 更新新手引导进度
func (ns *NewbieSys) UpdateGuideProgress(guideId uint32, completed bool) {
	if ns.newbieData.GuideProgress == nil {
		ns.newbieData.GuideProgress = make(map[uint32]bool)
	}
	ns.newbieData.GuideProgress[guideId] = completed

	// 检查是否所有引导都完成
	allCompleted := true
	for _, completed := range ns.newbieData.GuideProgress {
		if !completed {
			allCompleted = false
			break
		}
	}
	if allCompleted {
		ns.newbieData.GuideCompleted = true
	}
}

// IsGuideCompleted 检查新手引导是否完成
func (ns *NewbieSys) IsGuideCompleted() bool {
	return ns.newbieData.GuideCompleted
}

// GetGuideProgress 获取新手引导进度
func (ns *NewbieSys) GetGuideProgress(guideId uint32) bool {
	if ns.newbieData.GuideProgress == nil {
		return false
	}
	return ns.newbieData.GuideProgress[guideId]
}

// GetNewbieData 获取新手保护数据
func (ns *NewbieSys) GetNewbieData() *protocol.SiNewbieData {
	return ns.newbieData
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysNewbie), func() iface.ISystem {
		return NewNewbieSys()
	})
}
