// pathfinding.go 实现寻路算法：A*和直线寻路
package entitysystem

import (
	"container/heap"
	"math"
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

// PathfindingType 寻路算法类型（与jsonconf中的定义保持一致）
type PathfindingType = jsonconf.PathfindingType

const (
	PathfindingTypeStraight = jsonconf.PathfindingTypeStraight
	PathfindingTypeAStar    = jsonconf.PathfindingTypeAStar
)

// PathfindingResult 寻路结果
type PathfindingResult struct {
	Path     []*argsdef.Position // 路径点列表（不包括起点，包括终点）
	Found    bool                // 是否找到路径
	PathType PathfindingType     // 使用的寻路算法
}

// FindPath 根据寻路算法类型查找路径
// 注意：所有坐标参数都是格子坐标（不是像素坐标）
// fromX, fromY: 起点格子坐标
// toX, toY: 终点格子坐标
// 返回的路径点也是格子坐标
func FindPath(scene iface.IScene, fromX, fromY, toX, toY uint32, pathType PathfindingType) *PathfindingResult {
	if scene == nil {
		return &PathfindingResult{Found: false}
	}

	// 检查起点和终点是否可行走（格子坐标）
	if !scene.IsWalkable(int(fromX), int(fromY)) {
		return &PathfindingResult{Found: false}
	}
	if !scene.IsWalkable(int(toX), int(toY)) {
		return &PathfindingResult{Found: false}
	}

	switch pathType {
	case PathfindingTypeStraight:
		return findStraightPath(scene, fromX, fromY, toX, toY)
	case PathfindingTypeAStar:
		return findAStarPath(scene, fromX, fromY, toX, toY)
	default:
		return &PathfindingResult{Found: false}
	}
}

// findStraightPath 直线寻路：尝试走最短直线，遇到障碍物时绕过
// 注意：所有坐标参数都是格子坐标，返回的路径点也是格子坐标
func findStraightPath(scene iface.IScene, fromX, fromY, toX, toY uint32) *PathfindingResult {
	result := &PathfindingResult{
		Path:     make([]*argsdef.Position, 0),
		PathType: PathfindingTypeStraight,
	}

	dx := int64(toX) - int64(fromX)
	dy := int64(toY) - int64(fromY)
	dist := math.Sqrt(float64(dx*dx + dy*dy))

	if dist < 1.0 {
		// 起点和终点相同或非常接近
		result.Found = true
		result.Path = append(result.Path, &argsdef.Position{X: toX, Y: toY})
		return result
	}

	// 使用改进的直线寻路：遇到障碍时尝试绕过
	currentX, currentY := int32(fromX), int32(fromY)
	targetX, targetY := int32(toX), int32(toY)
	visited := make(map[int32]bool) // 防止重复访问
	maxSteps := int(dist * 2)       // 允许最多绕行一定距离
	if maxSteps > 500 {
		maxSteps = 500
	}

	// 8方向移动（优先保持直线方向）
	dirs := [][]int32{
		{0, 1}, {1, 0}, {0, -1}, {-1, 0}, // 上下左右
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1}, // 对角线
	}

	getDirPriority := func(dirX, dirY int32) float64 {
		// 计算方向与目标方向的夹角，优先选择更接近目标方向的
		dirToTargetX := float64(targetX - currentX)
		dirToTargetY := float64(targetY - currentY)
		dirLen := math.Sqrt(dirToTargetX*dirToTargetX + dirToTargetY*dirToTargetY)
		if dirLen < 0.1 {
			return 0 // 已到达目标
		}
		dirToTargetX /= dirLen
		dirToTargetY /= dirLen

		dirVecX := float64(dirX)
		dirVecY := float64(dirY)
		dirVecLen := math.Sqrt(dirVecX*dirVecX + dirVecY*dirVecY)
		if dirVecLen < 0.1 {
			return 1000
		}
		dirVecX /= dirVecLen
		dirVecY /= dirVecLen

		// 点积，值越大越接近目标方向
		dot := dirToTargetX*dirVecX + dirToTargetY*dirVecY
		return 1.0 - dot // 转换为优先级（越小越优先）
	}

	step := 0
	for step < maxSteps {
		step++

		// 检查是否到达目标
		if currentX == targetX && currentY == targetY {
			if scene.IsWalkable(int(currentX), int(currentY)) {
				result.Found = true
				// 添加终点（如果不在路径中）
				if len(result.Path) == 0 || result.Path[len(result.Path)-1].X != uint32(targetX) || result.Path[len(result.Path)-1].Y != uint32(targetY) {
					result.Path = append(result.Path, &argsdef.Position{X: uint32(targetX), Y: uint32(targetY)})
				}
			}
			break
		}

		// 计算理想的下一个点（直线方向）
		idealNextX := currentX
		idealNextY := currentY
		if dx != 0 || dy != 0 {
			stepSize := int32(1)
			if math.Abs(float64(dx)) > math.Abs(float64(dy)) {
				if dx > 0 {
					idealNextX = currentX + stepSize
				} else {
					idealNextX = currentX - stepSize
				}
				idealNextY = currentY + int32(float64(dy)/math.Abs(float64(dx))*float64(stepSize))
			} else {
				if dy > 0 {
					idealNextY = currentY + stepSize
				} else {
					idealNextY = currentY - stepSize
				}
				idealNextX = currentX + int32(float64(dx)/math.Abs(float64(dy))*float64(stepSize))
			}
		}

		// 优先尝试理想方向
		nextX, nextY := idealNextX, idealNextY
		found := false

		if scene.IsWalkable(int(idealNextX), int(idealNextY)) {
			key := idealNextY*10000 + idealNextX
			if !visited[key] {
				nextX, nextY = idealNextX, idealNextY
				found = true
			}
		}

		// 如果理想方向不可行，尝试其他方向（优先选择更接近目标方向的）
		if !found {
			type dirOption struct {
				dirX, dirY int32
				priority   float64
			}
			options := make([]dirOption, 0, 8)
			for _, dir := range dirs {
				nx := currentX + dir[0]
				ny := currentY + dir[1]
				if scene.IsWalkable(int(nx), int(ny)) {
					key := ny*10000 + nx
					if !visited[key] {
						priority := getDirPriority(dir[0], dir[1])
						options = append(options, dirOption{dirX: dir[0], dirY: dir[1], priority: priority})
					}
				}
			}

			// 按优先级排序
			for i := 0; i < len(options)-1; i++ {
				for j := i + 1; j < len(options); j++ {
					if options[i].priority > options[j].priority {
						options[i], options[j] = options[j], options[i]
					}
				}
			}

			// 选择优先级最高的
			if len(options) > 0 {
				nextX = currentX + options[0].dirX
				nextY = currentY + options[0].dirY
				found = true
			}
		}

		if !found {
			// 无法继续前进
			break
		}

		// 添加到路径
		key := nextY*10000 + nextX
		visited[key] = true
		result.Path = append(result.Path, &argsdef.Position{X: uint32(nextX), Y: uint32(nextY)})
		currentX, currentY = nextX, nextY

		// 更新到目标的方向向量
		dx = int64(targetX) - int64(currentX)
		dy = int64(targetY) - int64(currentY)
	}

	if len(result.Path) > 0 {
		result.Found = true
		// 确保终点在路径中
		lastPos := result.Path[len(result.Path)-1]
		if lastPos.X != uint32(targetX) || lastPos.Y != uint32(targetY) {
			if scene.IsWalkable(int(targetX), int(targetY)) {
				result.Path = append(result.Path, &argsdef.Position{X: uint32(targetX), Y: uint32(targetY)})
			}
		}
	}

	return result
}

// A*寻路算法相关结构
type astarNode struct {
	x, y    int32
	g, h, f float64
	parent  *astarNode
	index   int
}

type astarPriorityQueue []*astarNode

func (pq astarPriorityQueue) Len() int { return len(pq) }

func (pq astarPriorityQueue) Less(i, j int) bool {
	return pq[i].f < pq[j].f
}

func (pq astarPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *astarPriorityQueue) Push(x interface{}) {
	n := len(*pq)
	node := x.(*astarNode)
	node.index = n
	*pq = append(*pq, node)
}

func (pq *astarPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil
	node.index = -1
	*pq = old[0 : n-1]
	return node
}

// findAStarPath A*寻路算法
// 注意：所有坐标参数都是格子坐标，返回的路径点也是格子坐标
// 距离计算使用格子距离（曼哈顿距离或欧几里得距离）
func findAStarPath(scene iface.IScene, fromX, fromY, toX, toY uint32) *PathfindingResult {
	result := &PathfindingResult{
		Path:     make([]*argsdef.Position, 0),
		PathType: PathfindingTypeAStar,
	}

	startX, startY := int32(fromX), int32(fromY)
	endX, endY := int32(toX), int32(toY)

	// 如果起点和终点相同
	if startX == endX && startY == endY {
		result.Found = true
		result.Path = append(result.Path, &argsdef.Position{X: toX, Y: toY})
		return result
	}

	// 启发式函数：曼哈顿距离（格子距离）
	heuristic := func(x1, y1, x2, y2 int32) float64 {
		dx := math.Abs(float64(x1 - x2))
		dy := math.Abs(float64(y1 - y2))
		return dx + dy // 格子距离
	}

	// 初始化
	openSet := make(astarPriorityQueue, 0)
	heap.Init(&openSet)
	closedSet := make(map[int32]bool)
	nodeMap := make(map[int32]*astarNode)

	startNode := &astarNode{
		x:      startX,
		y:      startY,
		g:      0,
		h:      heuristic(startX, startY, endX, endY),
		parent: nil,
	}
	startNode.f = startNode.g + startNode.h
	startKey := startY*10000 + startX // 简单的key生成
	nodeMap[startKey] = startNode
	heap.Push(&openSet, startNode)

	// 8方向移动
	dirs := [][]int32{
		{0, 1}, {1, 0}, {0, -1}, {-1, 0}, // 上下左右
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1}, // 对角线
	}
	dirCosts := []float64{1.0, 1.0, 1.0, 1.0, 1.414, 1.414, 1.414, 1.414}

	maxIterations := 2000
	iterations := 0

	for openSet.Len() > 0 && iterations < maxIterations {
		iterations++
		current := heap.Pop(&openSet).(*astarNode)

		// 检查是否到达终点
		if current.x == endX && current.y == endY {
			// 重构路径
			path := make([]*argsdef.Position, 0)
			node := current
			for node != nil {
				path = append(path, &argsdef.Position{X: uint32(node.x), Y: uint32(node.y)})
				node = node.parent
			}
			// 反转路径（从起点到终点）
			for i := len(path) - 2; i >= 0; i-- {
				result.Path = append(result.Path, path[i])
			}
			result.Found = true
			return result
		}

		currentKey := current.y*10000 + current.x
		closedSet[currentKey] = true

		// 检查邻居节点
		for i, dir := range dirs {
			nx := current.x + dir[0]
			ny := current.y + dir[1]

			// 边界检查
			if nx < 0 || ny < 0 {
				continue
			}

			// 检查是否可行走
			if !scene.IsWalkable(int(nx), int(ny)) {
				continue
			}

			neighborKey := ny*10000 + nx
			if closedSet[neighborKey] {
				continue
			}

			// 计算新的g值
			newG := current.g + dirCosts[i]

			neighbor, exists := nodeMap[neighborKey]
			if !exists {
				neighbor = &astarNode{
					x:      nx,
					y:      ny,
					g:      newG,
					h:      heuristic(nx, ny, endX, endY),
					parent: current,
				}
				neighbor.f = neighbor.g + neighbor.h
				nodeMap[neighborKey] = neighbor
				heap.Push(&openSet, neighbor)
			} else if newG < neighbor.g {
				// 找到更优路径
				neighbor.g = newG
				neighbor.f = neighbor.g + neighbor.h
				neighbor.parent = current
				heap.Fix(&openSet, neighbor.index)
			}
		}
	}

	// 未找到路径
	result.Found = false
	return result
}
