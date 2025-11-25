package systems

import (
	"time"

	"postapocgame/server/example/internal/client"
	"postapocgame/server/internal/protocol"
)

type GMSystem struct {
	core *client.Core
}

func NewGMSystem(core *client.Core) *GMSystem {
	return &GMSystem{core: core}
}

func (s *GMSystem) Exec(name string, args []string, timeout time.Duration) (*protocol.S2CGMCommandResultReq, error) {
	if err := s.core.SendGMCommand(name, args); err != nil {
		return nil, err
	}
	return s.core.WaitGMResult(timeout)
}
