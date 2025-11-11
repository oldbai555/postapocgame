/**
 * @Author: zjj
 * @Date: 2025/11/6
 * @Desc:
**/

package argsdef

// SessionInfo 会话信息
type SessionInfo struct {
	SessionId string
	RoleId    uint64
	CreatedAt int64
}

func (s *SessionInfo) SetRoleId(roleId uint64) {
	s.RoleId = roleId
}

func (s *SessionInfo) GetRoleId() uint64 {
	return s.RoleId
}
func (s *SessionInfo) GetSessionId() string {
	return s.SessionId
}
