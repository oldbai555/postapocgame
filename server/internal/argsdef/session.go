/**
 * @Author: zjj
 * @Date: 2025/11/6
 * @Desc:
**/

package argsdef

// SessionInfo 会话信息
type SessionInfo struct {
	SessionId string
	AccountID uint // 账号ID
	RoleId    uint64
	Token     string // 登录token
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

func (s *SessionInfo) SetAccountID(accountID uint) {
	s.AccountID = accountID
}

func (s *SessionInfo) GetAccountID() uint {
	return s.AccountID
}

func (s *SessionInfo) SetToken(token string) {
	s.Token = token
}

func (s *SessionInfo) GetToken() string {
	return s.Token
}
