package systems

import (
	"context"

	"postapocgame/server/example/internal/client"
)

type MoveSystem struct {
	runner *client.MoveRunner
}

func NewMoveSystem(core *client.Core) *MoveSystem {
	return &MoveSystem{
		runner: core.MoveRunner(),
	}
}

func (s *MoveSystem) MoveDelta(ctx context.Context, dx, dy int32, cb *client.MoveCallbacks) error {
	return s.runner.MoveBy(ctx, dx, dy, cb)
}

func (s *MoveSystem) MoveTo(ctx context.Context, tileX, tileY uint32, cb *client.MoveCallbacks) error {
	return s.runner.MoveTo(ctx, tileX, tileY, cb)
}

func (s *MoveSystem) Resume(ctx context.Context, cb *client.MoveCallbacks) error {
	return s.runner.Resume(ctx, cb)
}
