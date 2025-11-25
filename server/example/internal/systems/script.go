package systems

import (
	"context"
	"time"
)

type ScriptSystem struct {
	move   *MoveSystem
	scene  *SceneSystem
	combat *CombatSystem
}

func NewScriptSystem(move *MoveSystem, scene *SceneSystem, combat *CombatSystem) *ScriptSystem {
	return &ScriptSystem{
		move:   move,
		scene:  scene,
		combat: combat,
	}
}

// RunDemo 在当前场景完成一个矩形巡逻，并尝试攻击第一个可见实体。
func (s *ScriptSystem) RunDemo(ctx context.Context) error {
	status := s.scene.Status()
	currentX := status.PosX
	currentY := status.PosY

	// 简单的矩形巡逻
	offsets := [][2]int32{
		{2, 0},
		{0, 2},
		{-2, 0},
		{0, -2},
	}

	for _, offset := range offsets {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		targetX := clampTileCoord(int32(currentX) + offset[0])
		targetY := clampTileCoord(int32(currentY) + offset[1])
		if err := s.move.MoveTo(ctx, targetX, targetY, nil); err != nil {
			return err
		}
		currentX = targetX
		currentY = targetY
	}

	// 如果视野中有实体，尝试普通攻击第一个
	entities := s.scene.ObservedEntities()
	if len(entities) > 0 {
		_, _ = s.combat.NormalAttack(entities[0].Handle, 2*time.Second)
	}
	return nil
}

func clampTileCoord(v int32) uint32 {
	if v < 0 {
		return 0
	}
	return uint32(v)
}
