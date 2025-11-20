package entitysystem

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
	"time"
)

const (
	// 操作频率限制（每10秒最多100次操作）
	maxOperationsPer10Seconds = 100
	operationWindowSeconds    = 10

	// 可疑行为阈值（达到此值触发警告）
	suspiciousThreshold = 10

	// 每日重置时间（凌晨0点）
	resetHour = 0
)

// AntiCheatSys 防作弊系统
type AntiCheatSys struct {
	*BaseSystem
	antiCheatData *protocol.SiAntiCheatData
}

// NewAntiCheatSys 创建防作弊系统
func NewAntiCheatSys() *AntiCheatSys {
	sys := &AntiCheatSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysAntiCheat)),
	}
	return sys
}

func GetAntiCheatSys(ctx context.Context) *AntiCheatSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysAntiCheat))
	if system == nil {
		log.Errorf("not found system [%v] error:%v", protocol.SystemId_SysAntiCheat, err)
		return nil
	}
	sys := system.(*AntiCheatSys)
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error:%v", protocol.SystemId_SysAntiCheat, err)
		return nil
	}
	return sys
}

// OnInit 系统初始化
func (acs *AntiCheatSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("anti cheat sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果anti_cheat_data不存在，则初始化
	if binaryData.AntiCheatData == nil {
		binaryData.AntiCheatData = &protocol.SiAntiCheatData{
			OperationCount:    make(map[string]int32),
			LastOperationTime: make(map[string]int64),
			SuspiciousCount:   0,
			LastResetTime:     servertime.Now().Unix(),
			IsBanned:          false,
			BanEndTime:        0,
		}
	}
	acs.antiCheatData = binaryData.AntiCheatData

	// 检查是否需要每日重置
	acs.checkAndResetDaily()

	// 检查封禁状态
	acs.checkBanStatus()
}

// checkAndResetDaily 检查并执行每日重置
func (acs *AntiCheatSys) checkAndResetDaily() {
	if acs.antiCheatData == nil {
		return
	}

	now := servertime.Now()
	lastReset := time.Unix(acs.antiCheatData.LastResetTime, 0)

	// 如果已经过了凌晨0点，重置计数
	if now.Day() != lastReset.Day() || now.Sub(lastReset) > 24*time.Hour {
		acs.antiCheatData.OperationCount = make(map[string]int32)
		acs.antiCheatData.LastOperationTime = make(map[string]int64)
		acs.antiCheatData.SuspiciousCount = 0
		acs.antiCheatData.LastResetTime = now.Unix()

		log.Infof("[AntiCheatSys] Daily reset completed")
	}
}

// checkBanStatus 检查封禁状态
func (acs *AntiCheatSys) checkBanStatus() {
	if acs.antiCheatData == nil {
		return
	}

	if !acs.antiCheatData.IsBanned {
		return
	}

	// 检查封禁是否已过期
	if acs.antiCheatData.BanEndTime > 0 {
		now := servertime.Now().Unix()
		if now >= acs.antiCheatData.BanEndTime {
			// 封禁已过期，解除封禁
			acs.antiCheatData.IsBanned = false
			acs.antiCheatData.BanEndTime = 0
			log.Infof("[AntiCheatSys] Ban expired, unbanning player")
		}
	}
}

// CheckOperationFrequency 检查操作频率（返回是否允许操作）
func (acs *AntiCheatSys) CheckOperationFrequency(operationType string) bool {
	if acs.antiCheatData == nil {
		return true
	}

	// 检查是否被封禁
	if acs.antiCheatData.IsBanned {
		return false
	}

	now := servertime.Now().Unix()

	// 获取操作计数和最后操作时间
	count := acs.antiCheatData.OperationCount[operationType]
	lastTime := acs.antiCheatData.LastOperationTime[operationType]

	// 如果距离上次操作超过窗口时间，重置计数
	if now-lastTime > operationWindowSeconds {
		count = 0
	}

	// 检查操作频率
	if count >= maxOperationsPer10Seconds {
		// 操作过于频繁，增加可疑行为计数
		acs.antiCheatData.SuspiciousCount++
		log.Warnf("[AntiCheatSys] Operation frequency exceeded: type=%s, count=%d", operationType, count)
		return false
	}

	// 更新计数和时间
	acs.antiCheatData.OperationCount[operationType] = count + 1
	acs.antiCheatData.LastOperationTime[operationType] = now

	return true
}

// RecordSuspiciousBehavior 记录可疑行为
func (acs *AntiCheatSys) RecordSuspiciousBehavior(reason string) {
	if acs.antiCheatData == nil {
		return
	}

	acs.antiCheatData.SuspiciousCount++
	log.Warnf("[AntiCheatSys] Suspicious behavior detected: reason=%s, count=%d", reason, acs.antiCheatData.SuspiciousCount)

	// 如果可疑行为过多，临时封禁（1小时）
	if acs.antiCheatData.SuspiciousCount >= suspiciousThreshold*2 {
		acs.BanPlayer(1*time.Hour, "Too many suspicious behaviors")
	} else if acs.antiCheatData.SuspiciousCount >= suspiciousThreshold {
		log.Warnf("[AntiCheatSys] Player reached suspicious threshold: count=%d", acs.antiCheatData.SuspiciousCount)
	}
}

// BanPlayer 封禁玩家
func (acs *AntiCheatSys) BanPlayer(duration time.Duration, reason string) {
	if acs.antiCheatData == nil {
		return
	}

	acs.antiCheatData.IsBanned = true
	if duration > 0 {
		acs.antiCheatData.BanEndTime = servertime.Now().Add(duration).Unix()
	} else {
		// 永久封禁
		acs.antiCheatData.BanEndTime = 0
	}

	log.Warnf("[AntiCheatSys] Player banned: duration=%v, reason=%s", duration, reason)
}

// UnbanPlayer 解封玩家
func (acs *AntiCheatSys) UnbanPlayer() {
	if acs.antiCheatData == nil {
		return
	}

	acs.antiCheatData.IsBanned = false
	acs.antiCheatData.BanEndTime = 0
	log.Infof("[AntiCheatSys] Player unbanned")
}

// IsBanned 检查是否被封禁
func (acs *AntiCheatSys) IsBanned() bool {
	if acs.antiCheatData == nil {
		return false
	}

	// 检查封禁状态
	acs.checkBanStatus()
	return acs.antiCheatData.IsBanned
}

// GetSuspiciousCount 获取可疑行为计数
func (acs *AntiCheatSys) GetSuspiciousCount() int32 {
	if acs.antiCheatData == nil {
		return 0
	}
	return acs.antiCheatData.SuspiciousCount
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysAntiCheat), func() iface.ISystem {
		return NewAntiCheatSys()
	})
}
