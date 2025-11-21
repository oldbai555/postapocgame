package entitysystem

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/iface"
)

// GuildSys 公会系统
type GuildSys struct {
	*BaseSystem
	data *protocol.SiGuildData
}

// NewGuildSys 创建公会系统
func NewGuildSys() iface.ISystem {
	return &GuildSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysGuild)),
	}
}

func (s *GuildSys) OnInit(ctx context.Context) {
	role, err := GetIPlayerRoleByContext(ctx)
	if err != nil || role == nil {
		return
	}
	bd := role.GetBinaryData()
	if bd.GuildData == nil {
		bd.GuildData = &protocol.SiGuildData{
			GuildId:  0,
			Position: 0,
			JoinTime: 0,
		}
	}
	s.data = bd.GuildData
}

// GetGuildId 获取公会ID
func (s *GuildSys) GetGuildId() uint64 {
	if s.data == nil {
		return 0
	}
	return s.data.GuildId
}

// SetGuildId 设置公会ID
func (s *GuildSys) SetGuildId(guildId uint64) {
	if s.data != nil {
		s.data.GuildId = guildId
	}
}

// GetPosition 获取职位
func (s *GuildSys) GetPosition() uint32 {
	if s.data == nil {
		return 0
	}
	return s.data.Position
}

// SetPosition 设置职位
func (s *GuildSys) SetPosition(position uint32) {
	if s.data != nil {
		s.data.Position = position
	}
}

// GetJoinTime 获取加入时间
func (s *GuildSys) GetJoinTime() int64 {
	if s.data == nil {
		return 0
	}
	return s.data.JoinTime
}

// SetJoinTime 设置加入时间
func (s *GuildSys) SetJoinTime(joinTime int64) {
	if s.data != nil {
		s.data.JoinTime = joinTime
	}
}

// GetGuildSys 获取公会系统
func GetGuildSys(ctx context.Context) *GuildSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysGuild))
	if system == nil {
		return nil
	}
	sys := system.(*GuildSys)
	if sys == nil || !sys.IsOpened() {
		return nil
	}
	return sys
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysGuild), NewGuildSys)
}
