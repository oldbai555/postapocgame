package systems

import (
	"time"

	"postapocgame/server/example/internal/client"
	"postapocgame/server/internal/protocol"
)

type DungeonSystem struct {
	core *client.Core
}

func NewDungeonSystem(core *client.Core) *DungeonSystem {
	return &DungeonSystem{core: core}
}

func (s *DungeonSystem) Enter(dungeonID, difficulty uint32, timeout time.Duration) (*protocol.S2CEnterDungeonResultReq, error) {
	if err := s.core.EnterDungeonReq(dungeonID, difficulty); err != nil {
		return nil, err
	}
	return s.core.WaitEnterDungeonResult(timeout)
}
