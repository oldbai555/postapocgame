package systems

import (
	"time"

	"postapocgame/server/example/internal/client"
)

type CombatSystem struct {
	core *client.Core
}

func NewCombatSystem(core *client.Core) *CombatSystem {
	return &CombatSystem{core: core}
}

func (s *CombatSystem) NormalAttack(target uint64, wait time.Duration) error {
	if err := s.core.CastNormalAttack(target); err != nil {
		return err
	}
	_ = wait
	return nil
}
