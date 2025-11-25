package client

import (
	"container/heap"
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/pkg/log"
)

var (
	ErrMoveBlocked      = errors.New("move runner: target not reachable")
	ErrMoveCancelled    = errors.New("move runner: cancelled")
	ErrMoveOutOfSync    = errors.New("move runner: server position not in sync")
	errNoSceneMap       = errors.New("move runner: scene map not loaded")
	errTargetUnwalkable = errors.New("move runner: target not walkable")
	errNoLastTarget     = errors.New("move runner: no previous target")
)

type MoveRunner struct {
	core *Core

	mu         sync.Mutex
	cancelFn   context.CancelFunc
	moveToken  int64
	lastTarget *tilePoint

	options MoveRunnerOptions
}

type MoveRunnerOptions struct {
	Speed                uint32
	SyncTimeout          time.Duration
	PollInterval         time.Duration
	MaxRetry             int
	PositionTolerance    uint32
	LatencyToleranceMs   int64
	SpeedToleranceFactor float64
}

type MoveCallbacks struct {
	OnArrived func()
	OnAbort   func(error)
}

func NewMoveRunner(core *Core) *MoveRunner {
	return &MoveRunner{
		core: core,
		options: MoveRunnerOptions{
			Speed:                defaultMoveSpeed,
			SyncTimeout:          1200 * time.Millisecond,
			PollInterval:         60 * time.Millisecond,
			MaxRetry:             2,
			PositionTolerance:    1,
			LatencyToleranceMs:   200,
			SpeedToleranceFactor: 1.5,
		},
	}
}

func (r *MoveRunner) MoveBy(ctx context.Context, deltaX, deltaY int32, cb *MoveCallbacks) error {
	status := r.core.RoleStatus()
	targetX := clampTile(int64(status.PosX) + int64(deltaX))
	targetY := clampTile(int64(status.PosY) + int64(deltaY))
	return r.MoveTo(ctx, targetX, targetY, cb)
}

func (r *MoveRunner) MoveTo(ctx context.Context, tileX, tileY uint32, cb *MoveCallbacks) error {
	r.mu.Lock()
	if r.cancelFn != nil {
		r.cancelFn()
	}
	runCtx, cancel := context.WithCancel(ctx)
	r.cancelFn = cancel
	r.moveToken++
	token := r.moveToken
	targetCopy := &tilePoint{X: tileX, Y: tileY}
	r.lastTarget = targetCopy
	r.mu.Unlock()

	defer func(tok int64) {
		r.mu.Lock()
		if r.moveToken == tok {
			r.cancelFn = nil
		}
		r.mu.Unlock()
	}(token)

	err := r.run(runCtx, tileX, tileY)
	if err != nil {
		if cb != nil && cb.OnAbort != nil {
			cb.OnAbort(err)
		}
		return err
	}
	if cb != nil && cb.OnArrived != nil {
		cb.OnArrived()
	}
	return nil
}

func (r *MoveRunner) Resume(ctx context.Context, cb *MoveCallbacks) error {
	r.mu.Lock()
	target := r.lastTarget
	r.mu.Unlock()
	if target == nil {
		return errNoLastTarget
	}
	return r.MoveTo(ctx, target.X, target.Y, cb)
}

func (r *MoveRunner) run(ctx context.Context, tileX, tileY uint32) error {
	sceneMap := r.core.CurrentSceneMap()
	if sceneMap == nil {
		return errNoSceneMap
	}
	if !sceneMap.IsWalkable(int32(tileX), int32(tileY)) {
		return errTargetUnwalkable
	}

	retry := 0
	for {
		select {
		case <-ctx.Done():
			return ErrMoveCancelled
		default:
		}

		status := r.core.RoleStatus()
		start := tilePoint{X: status.PosX, Y: status.PosY}
		target := tilePoint{X: tileX, Y: tileY}
		if start == target {
			return nil
		}

		path, err := r.makePath(sceneMap, start, target)
		if err != nil {
			return err
		}
		log.Infof("[%s] MoveRunner start from (%d,%d) -> (%d,%d), path=%s",
			r.core.GetPlayerID(), start.X, start.Y, target.X, target.Y, describePath(path))

		err = r.executePath(ctx, path)
		if err == nil {
			return nil
		}
		if errors.Is(err, ErrMoveOutOfSync) && retry < r.options.MaxRetry {
			retry++
			continue
		}
		return err
	}
}

func (r *MoveRunner) executePath(ctx context.Context, path []tilePoint) error {
	current := path[0]
	for i := 1; i < len(path); i++ {
		next := path[i]
		select {
		case <-ctx.Done():
			return ErrMoveCancelled
		default:
		}

		log.Debugf("[%s] MoveRunner segment %d/%d: (%d,%d) -> (%d,%d)",
			r.core.GetPlayerID(), i, len(path)-1, current.X, current.Y, next.X, next.Y)

		if err := r.core.sendMoveChunk(current.X, current.Y, next.X, next.Y); err != nil {
			return err
		}

		r.core.updateLocalPosition(next.X, next.Y)

		if !r.waitForServerPos(ctx, next) {
			log.Warnf("[%s] MoveRunner wait server position timeout, target=(%d,%d)",
				r.core.GetPlayerID(), next.X, next.Y)
			return ErrMoveOutOfSync
		}
		current = next
	}
	return nil
}

func (r *MoveRunner) waitForServerPos(ctx context.Context, target tilePoint) bool {
	deadline := time.Now().Add(r.options.SyncTimeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		status := r.core.RoleStatus()
		if withinTolerance(tilePoint{status.PosX, status.PosY}, target, r.options.PositionTolerance) {
			return true
		}
		time.Sleep(r.options.PollInterval)
	}
	return false
}

func (r *MoveRunner) makePath(sceneMap *jsonconf.GameMap, start, target tilePoint) ([]tilePoint, error) {
	if r.hasLineOfSight(sceneMap, start, target) {
		return []tilePoint{start, target}, nil
	}
	return r.findPath(sceneMap, start, target)
}

func (r *MoveRunner) hasLineOfSight(sceneMap *jsonconf.GameMap, start, target tilePoint) bool {
	x0 := int32(start.X)
	y0 := int32(start.Y)
	x1 := int32(target.X)
	y1 := int32(target.Y)

	dx := abs32(x1 - x0)
	dy := -abs32(y1 - y0)
	sx := int32(1)
	if x0 >= x1 {
		sx = -1
	}
	sy := int32(1)
	if y0 >= y1 {
		sy = -1
	}
	err := dx + dy

	for {
		if !sceneMap.IsWalkable(x0, y0) {
			return false
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
	return true
}

func (r *MoveRunner) findPath(sceneMap *jsonconf.GameMap, start, target tilePoint) ([]tilePoint, error) {
	open := &tileHeap{}
	heap.Init(open)

	startNode := &pathNode{
		point: start,
	}
	heap.Push(open, startNode)

	cameFrom := make(map[tilePoint]*pathNode)
	gScore := map[tilePoint]float64{start: 0}
	fScore := map[tilePoint]float64{start: heuristic(start, target)}

	visited := make(map[tilePoint]bool)

	for open.Len() > 0 {
		current := heap.Pop(open).(*pathNode)
		if current.point == target {
			return reconstructPath(current), nil
		}
		if visited[current.point] {
			continue
		}
		visited[current.point] = true

		for _, neighbor := range neighbors(sceneMap, current.point) {
			tentativeG := gScore[current.point] + 1
			if val, ok := gScore[neighbor]; ok && tentativeG >= val {
				continue
			}
			gScore[neighbor] = tentativeG
			fScore[neighbor] = tentativeG + heuristic(neighbor, target)
			cameFrom[neighbor] = current
			heap.Push(open, &pathNode{
				point: neighbor,
				score: fScore[neighbor],
				prev:  current,
			})
		}
	}

	return nil, ErrMoveBlocked
}

func neighbors(sceneMap *jsonconf.GameMap, p tilePoint) []tilePoint {
	dirs := [][2]int32{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},
	}
	result := make([]tilePoint, 0, len(dirs))
	for _, d := range dirs {
		nx := int32(p.X) + d[0]
		ny := int32(p.Y) + d[1]
		if sceneMap.IsWalkable(nx, ny) {
			result = append(result, tilePoint{uint32(nx), uint32(ny)})
		}
	}
	return result
}

func reconstructPath(node *pathNode) []tilePoint {
	path := []tilePoint{}
	for node != nil {
		path = append([]tilePoint{node.point}, path...)
		node = node.prev
	}
	return path
}

type tilePoint struct {
	X uint32
	Y uint32
}

type pathNode struct {
	point tilePoint
	score float64
	prev  *pathNode
}

type tileHeap []*pathNode

func (h tileHeap) Len() int           { return len(h) }
func (h tileHeap) Less(i, j int) bool { return h[i].score < h[j].score }
func (h tileHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *tileHeap) Push(x interface{}) {
	*h = append(*h, x.(*pathNode))
}

func (h *tileHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func heuristic(a, b tilePoint) float64 {
	dx := float64(int64(a.X) - int64(b.X))
	dy := float64(int64(a.Y) - int64(b.Y))
	return math.Abs(dx) + math.Abs(dy)
}

func withinTolerance(a, b tilePoint, tol uint32) bool {
	dx := int64(a.X) - int64(b.X)
	dy := int64(a.Y) - int64(b.Y)
	return abs64Int(dx) <= int64(tol) && abs64Int(dy) <= int64(tol)
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func abs64Int(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

func describePath(path []tilePoint) string {
	if len(path) == 0 {
		return "[]"
	}
	const maxSteps = 6
	parts := make([]string, 0, len(path))
	for idx, p := range path {
		if idx >= maxSteps {
			parts = append(parts, "...")
			break
		}
		parts = append(parts, "("+
			strconv.FormatUint(uint64(p.X), 10)+","+
			strconv.FormatUint(uint64(p.Y), 10)+")")
	}
	return "[" + strings.Join(parts, "->") + "]"
}

func clampTile(v int64) uint32 {
	if v < 0 {
		return 0
	}
	if v > math.MaxInt32 {
		return uint32(math.MaxInt32)
	}
	return uint32(v)
}
