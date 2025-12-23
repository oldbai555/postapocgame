package systems

import "postapocgame/server/example/internal/client"

// Set 聚合了所有系统，面板只与系统交互
type Set struct {
	Account *AccountSystem
	Scene   *SceneSystem
	Move    *MoveSystem
	Combat  *CombatSystem
}

func NewSet(core *client.Core) *Set {
	moveSys := NewMoveSystem(core)
	sceneSys := NewSceneSystem(core)
	combatSys := NewCombatSystem(core)
	return &Set{
		Account: NewAccountSystem(core),
		Scene:   sceneSys,
		Move:    moveSys,
		Combat:  combatSys,
	}
}
