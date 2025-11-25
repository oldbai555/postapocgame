package systems

import (
	"postapocgame/server/example/internal/client"
)

type SceneSystem struct {
	core *client.Core
}

func NewSceneSystem(core *client.Core) *SceneSystem {
	return &SceneSystem{core: core}
}

func (s *SceneSystem) Status() client.RoleStatus {
	return s.core.RoleStatus()
}

func (s *SceneSystem) ObservedEntities() []*client.EntityView {
	return s.core.ObservedEntities()
}

func (s *SceneSystem) InScene() bool {
	return s.core.HasEnteredScene()
}
