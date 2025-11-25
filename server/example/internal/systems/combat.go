package systems

import (
	"time"

	"postapocgame/server/example/internal/client"
	"postapocgame/server/internal/protocol"
)

type CombatSystem struct {
	core *client.Core
}

func NewCombatSystem(core *client.Core) *CombatSystem {
	return &CombatSystem{core: core}
}

func (s *CombatSystem) NormalAttack(target uint64, wait time.Duration) (*protocol.SkillHitResultSt, error) {
	if err := s.core.CastNormalAttack(target); err != nil {
		return nil, err
	}
	return s.core.WaitForSkillResult(target, wait)
}
