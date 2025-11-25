package systems

import (
	"postapocgame/server/example/internal/client"
	"postapocgame/server/internal/protocol"
)

type AccountSystem struct {
	core *client.Core
}

func NewAccountSystem(core *client.Core) *AccountSystem {
	return &AccountSystem{core: core}
}

func (s *AccountSystem) Register(username, password string) error {
	return s.core.RegisterAccount(username, password)
}

func (s *AccountSystem) Login(username, password string) error {
	return s.core.LoginAccount(username, password)
}

func (s *AccountSystem) ListRoles() ([]*protocol.PlayerSimpleData, error) {
	return s.core.ListRoles()
}

func (s *AccountSystem) CreateRole(name string, job, sex uint32) error {
	return s.core.CreateRole(name, job, sex)
}

func (s *AccountSystem) EnterRole(roleID uint64) error {
	return s.core.EnterGame(roleID)
}
