package scene

import (
	"fmt"
	"math/rand"
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/entity"
	"postapocgame/server/service/dungeonserver/internel/entitymgr"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"sync"
)

// SceneSt 场景结构
type SceneSt struct {
	sceneId uint32
	fuBenId uint32
	name    string
	width   int
	height  int

	// 实体管理
	entities map[uint64]iface.IEntity
	entityMu sync.RWMutex

	// AOI管理器
	aoiMgr *AOIManager

	// 地图数据
	walkableMap [][]bool // 可行走地图

	// 怪物管理
	monsters  map[uint64]*entity.MonsterEntity
	monsterMu sync.RWMutex

	nextEntityId uint64
}

// NewSceneSt 创建场景
func NewSceneSt(sceneId, fuBenId uint32, name string, wIdth, height int) *SceneSt {
	scene := &SceneSt{
		sceneId:      sceneId,
		fuBenId:      fuBenId,
		name:         name,
		width:        wIdth,
		height:       height,
		entities:     make(map[uint64]iface.IEntity),
		monsters:     make(map[uint64]*entity.MonsterEntity),
		aoiMgr:       NewAOIManager(),
		nextEntityId: 1,
	}

	// 初始化地图
	scene.initMap()

	return scene
}

// initMap 初始化地图
func (s *SceneSt) initMap() {
	s.walkableMap = make([][]bool, s.height)
	for i := 0; i < s.height; i++ {
		s.walkableMap[i] = make([]bool, s.width)
		for j := 0; j < s.width; j++ {
			// 默认都可行走
			s.walkableMap[i][j] = true
		}
	}

	// 随机生成一些不可行走的点
	obstacleCount := (s.width * s.height) / 20 // 5%的地图不可行走
	for i := 0; i < obstacleCount; i++ {
		x := rand.Intn(s.width)
		y := rand.Intn(s.height)
		s.walkableMap[y][x] = false
	}

	log.Infof("Scene %d map initialized: %dx%d, obstacles: %d", s.sceneId, s.width, s.height, obstacleCount)
}

// IsWalkable 检查位置是否可行走
func (s *SceneSt) IsWalkable(x, y int) bool {
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return false
	}
	return s.walkableMap[y][x]
}

// GetRandomWalkablePos 获取随机可行走位置
func (s *SceneSt) GetRandomWalkablePos() (uint32, uint32) {
	maxAttempts := 100
	for i := 0; i < maxAttempts; i++ {
		x := rand.Intn(s.width)
		y := rand.Intn(s.height)
		if s.walkableMap[y][x] {
			return uint32(x), uint32(y)
		}
	}

	// 如果找不到，返回中心点
	return uint32(s.width / 2), uint32(s.height / 2)
}

// AddEntity 添加实体到场景
func (s *SceneSt) AddEntity(e iface.IEntity) error {
	s.entityMu.Lock()
	defer s.entityMu.Unlock()

	hdl := e.GetHdl()
	if _, exists := s.entities[hdl]; exists {
		return fmt.Errorf("entity already exists in scene: hdl=%d", hdl)
	}

	s.entities[hdl] = e
	e.SetSceneId(s.sceneId)
	e.SetFuBenId(s.fuBenId)

	// 注册到全局EntityMgr
	entityMgr := entitymgr.GetEntityMgr()
	if err := entityMgr.Register(e); err != nil {
		log.Warnf("Register entity to EntityMgr failed: %v", err)
	}

	// 添加到AOI管理器
	s.aoiMgr.AddEntity(e)

	// 触发进入场景回调
	e.OnEnterScene()

	log.Infof("Entity %d (hdl) entered scene %d", hdl, s.sceneId)

	return nil
}

// RemoveEntity 从场景移除实体
func (s *SceneSt) RemoveEntity(hdl uint64) error {
	s.entityMu.Lock()
	defer s.entityMu.Unlock()

	e, exists := s.entities[hdl]
	if !exists {
		return fmt.Errorf("entity not found in scene: hdl=%d", hdl)
	}

	// 从AOI管理器移除
	s.aoiMgr.RemoveEntity(e)

	// 触发离开场景回调
	e.OnLeaveScene()

	delete(s.entities, hdl)

	// 从全局EntityMgr注销
	entityMgr := entitymgr.GetEntityMgr()
	entityMgr.Unregister(hdl)

	log.Infof("Entity %d (hdl) left scene %d", hdl, s.sceneId)

	return nil
}

// GetEntity 获取实体
func (s *SceneSt) GetEntity(hdl uint64) (iface.IEntity, bool) {
	s.entityMu.RLock()
	defer s.entityMu.RUnlock()

	e, ok := s.entities[hdl]
	return e, ok
}

// GetAllEntities 获取所有实体
func (s *SceneSt) GetAllEntities() []iface.IEntity {
	s.entityMu.RLock()
	defer s.entityMu.RUnlock()

	entities := make([]iface.IEntity, 0, len(s.entities))
	for _, e := range s.entities {
		entities = append(entities, e)
	}
	return entities
}

// EntityMove 实体移动
func (s *SceneSt) EntityMove(hdl uint64, newX, newY uint32) error {
	e, ok := s.GetEntity(hdl)
	if !ok {
		return fmt.Errorf("entity not found: hdl=%d", hdl)
	}

	// 检查目标位置是否可行走
	if !s.IsWalkable(int(newX), int(newY)) {
		return fmt.Errorf("position not walkable: (%f, %f)", newX, newY)
	}

	oldPos := e.GetPosition()

	// 更新实体位置
	e.OnMove(newX, newY)

	// 更新AOI
	s.aoiMgr.UpdateEntity(e, oldPos, &argsdef.Position{X: newX, Y: newY})

	return nil
}

// SpawnMonster 生成怪物
func (s *SceneSt) SpawnMonster(monsterId uint32, name string, level uint32, x, y uint32) (*entity.MonsterEntity, error) {
	// 生成唯一hdl
	entityMgr := entitymgr.GetEntityMgr()
	hdl := entityMgr.GenHdl()

	// 创建怪物实体
	monster := entity.NewMonsterEntity(hdl, monsterId, name, level)
	monster.SetPosition(x, y)

	// 添加到场景
	if err := s.AddEntity(monster); err != nil {
		return nil, err
	}

	// 添加到怪物列表
	s.monsterMu.Lock()
	s.monsters[hdl] = monster
	s.monsterMu.Unlock()

	log.Infof("Monster spawned: hdl=%d, monsterId=%d, name=%s, pos=(%d,%d)",
		hdl, monsterId, name, x, y)

	return monster, nil
}

// RemoveMonster 移除怪物
func (s *SceneSt) RemoveMonster(hdl uint64) {
	s.monsterMu.Lock()
	delete(s.monsters, hdl)
	s.monsterMu.Unlock()

	s.RemoveEntity(hdl)
}

// GetMonsterCount 获取怪物数量
func (s *SceneSt) GetMonsterCount() int {
	s.monsterMu.RLock()
	defer s.monsterMu.RUnlock()
	return len(s.monsters)
}

// InitMonsters 初始化场景怪物
func (s *SceneSt) InitMonsters() {
	// 根据场景Id初始化不同的怪物
	switch s.sceneId {
	case 1:
		// 场景1是空白地图，无怪物
		log.Infof("Scene 1: Empty scene, no monsters")
	case 2:
		// 场景2：生成各种怪物
		s.spawnScene2Monsters()
	default:
		log.Infof("Scene %d: No monster configuration", s.sceneId)
	}
}

// spawnScene2Monsters 生成场景2的怪物
func (s *SceneSt) spawnScene2Monsters() {
	// 史莱姆 x10
	for i := 0; i < 10; i++ {
		x, y := s.GetRandomWalkablePos()
		s.SpawnMonster(10001, "史莱姆", 1, x, y)
	}

	// 哥布林 x8
	for i := 0; i < 8; i++ {
		x, y := s.GetRandomWalkablePos()
		s.SpawnMonster(10002, "哥布林", 5, x, y)
	}

	// 森林狼 x5
	for i := 0; i < 5; i++ {
		x, y := s.GetRandomWalkablePos()
		s.SpawnMonster(10003, "森林狼", 8, x, y)
	}

	// 哥布林首领 x2
	for i := 0; i < 2; i++ {
		x, y := s.GetRandomWalkablePos()
		s.SpawnMonster(20001, "哥布林首领", 10, x, y)
	}

	// 森林守护者 x1 (BOSS)
	x, y := s.GetRandomWalkablePos()
	s.SpawnMonster(30001, "森林守护者", 15, x, y)

	log.Infof("Scene 2 monsters spawned: total=%d", s.GetMonsterCount())
}

// GetSceneId 获取场景Id
func (s *SceneSt) GetSceneId() uint32 {
	return s.sceneId
}

// GetFuBenId 获取副本Id
func (s *SceneSt) GetFuBenId() uint32 {
	return s.fuBenId
}

// GetName 获取场景名称
func (s *SceneSt) GetName() string {
	return s.name
}

// GetWidth 获取场景宽度
func (s *SceneSt) GetWidth() int {
	return s.width
}

// GetHeight 获取场景高度
func (s *SceneSt) GetHeight() int {
	return s.height
}
