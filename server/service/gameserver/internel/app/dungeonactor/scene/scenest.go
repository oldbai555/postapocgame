package scene

import (
	"fmt"
	"math/rand"
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entity"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entitymgr"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/iface"
)

// SceneSt 场景结构
type SceneSt struct {
	sceneId  uint32
	fuBenId  uint32
	name     string
	width    int
	height   int
	bornArea *jsonconf.BornArea

	fuBen iface.IFuBen

	// 实体管理
	entities map[uint64]iface.IEntity

	// AOI管理器
	aoiMgr *AOIManager

	// 地图数据
	gameMap     *jsonconf.GameMap
	walkableMap [][]bool // fallback 使用

	// 怪物管理
	monsters map[uint64]*entity.MonsterEntity

	nextEntityId uint64
}

// NewSceneSt 创建场景
func NewSceneSt(fuBen iface.IFuBen, sceneId, fuBenId uint32, name string, wIdth, height int, mapData *jsonconf.GameMap, bornArea *jsonconf.BornArea) *SceneSt {
	scene := &SceneSt{
		sceneId:      sceneId,
		fuBenId:      fuBenId,
		name:         name,
		width:        wIdth,
		height:       height,
		fuBen:        fuBen,
		entities:     make(map[uint64]iface.IEntity),
		monsters:     make(map[uint64]*entity.MonsterEntity),
		aoiMgr:       NewAOIManager(),
		nextEntityId: 1,
		gameMap:      mapData,
		bornArea:     bornArea,
	}

	// 初始化地图
	scene.initMap()

	return scene
}

func (s *SceneSt) GetFuBen() iface.IFuBen {
	return s.fuBen
}

// initMap 初始化地图
func (s *SceneSt) initMap() {
	if s.gameMap != nil {
		s.width = int(s.gameMap.Width())
		s.height = int(s.gameMap.Height())
		log.Infof("Scene %d map loaded from config: %dx%d, movable=%d", s.sceneId, s.width, s.height, s.gameMap.MovableCount())
		return
	}

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
// 注意：x, y 是格子坐标（不是像素坐标）
func (s *SceneSt) IsWalkable(x, y int) bool {
	if s.gameMap != nil {
		return s.gameMap.IsWalkable(int32(x), int32(y))
	}
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return false
	}
	return s.walkableMap[y][x]
}

// GetRandomWalkablePos 获取随机可行走位置
// 返回：格子坐标（不是像素坐标）
func (s *SceneSt) GetRandomWalkablePos() (uint32, uint32) {
	if s.gameMap != nil {
		if x, y, ok := s.gameMap.RandomWalkableCoord(); ok {
			return uint32(x), uint32(y)
		}
		return uint32(s.width / 2), uint32(s.height / 2)
	}
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

// GetSpawnPos 获取出生点位置
// 返回：格子坐标（不是像素坐标）
func (s *SceneSt) GetSpawnPos() (uint32, uint32) {
	if x, y, ok := s.randomWalkableInBornArea(); ok {
		return x, y
	}
	return s.GetRandomWalkablePos()
}

func (s *SceneSt) randomWalkableInBornArea() (uint32, uint32, bool) {
	if s.bornArea == nil {
		return 0, 0, false
	}
	minX := clampInt(int(s.bornArea.X1), 0, s.width-1)
	maxX := clampInt(int(s.bornArea.X2), 0, s.width-1)
	minY := clampInt(int(s.bornArea.Y1), 0, s.height-1)
	maxY := clampInt(int(s.bornArea.Y2), 0, s.height-1)
	if maxX < minX || maxY < minY {
		return 0, 0, false
	}
	rangeX := maxX - minX + 1
	rangeY := maxY - minY + 1
	maxAttempts := rangeX * rangeY
	if maxAttempts < 16 {
		maxAttempts = 16
	}
	if maxAttempts > 256 {
		maxAttempts = 256
	}

	for i := 0; i < maxAttempts; i++ {
		x := minX + rand.Intn(rangeX)
		y := minY + rand.Intn(rangeY)
		if s.IsWalkable(x, y) {
			return uint32(x), uint32(y), true
		}
	}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if s.IsWalkable(x, y) {
				return uint32(x), uint32(y), true
			}
		}
	}
	return 0, 0, false
}

// AddEntity 添加实体到场景
func (s *SceneSt) AddEntity(e iface.IEntity) error {
	hdl := e.GetHdl()
	if _, exists := s.entities[hdl]; exists {
		return fmt.Errorf("entity already exists in scene: hdl=%d", hdl)
	}

	s.entities[hdl] = e
	e.SetSceneId(s.sceneId)
	e.SetFuBenId(s.fuBenId)

	if moveSys := e.GetMoveSys(); moveSys != nil {
		moveSys.BindScene(s)
	}

	// 注册到全局EntityMgr
	entityMgr := entitymgr.GetEntityMgr()
	if err := entityMgr.Register(e); err != nil {
		log.Warnf("Register entity to EntityMgr failed: %v", err)
	}
	entityMgr.BindScene(hdl, s)

	// 添加到AOI管理器
	s.aoiMgr.AddEntity(e)

	// 触发进入场景回调
	e.OnEnterScene()

	log.Infof("Entity %d (hdl) entered scene %d", hdl, s.sceneId)

	return nil
}

// RemoveEntity 从场景移除实体
func (s *SceneSt) RemoveEntity(hdl uint64) error {
	e, exists := s.entities[hdl]
	if !exists {
		return fmt.Errorf("entity not found in scene: hdl=%d", hdl)
	}

	// 从AOI管理器移除
	s.aoiMgr.RemoveEntity(e)

	if moveSys := e.GetMoveSys(); moveSys != nil {
		moveSys.UnbindScene(s)
	}

	// 触发离开场景回调
	e.OnLeaveScene()

	delete(s.entities, hdl)

	// 从全局EntityMgr注销
	entityMgr := entitymgr.GetEntityMgr()
	entityMgr.UnbindScene(hdl)
	entityMgr.Unregister(hdl)

	log.Infof("Entity %d (hdl) left scene %d", hdl, s.sceneId)

	return nil
}

// GetEntity 获取实体
func (s *SceneSt) GetEntity(hdl uint64) (iface.IEntity, bool) {
	e, ok := s.entities[hdl]
	return e, ok
}

// GetAllEntities 获取所有实体
func (s *SceneSt) GetAllEntities() []iface.IEntity {
	entities := make([]iface.IEntity, 0, len(s.entities))
	for _, e := range s.entities {
		entities = append(entities, e)
	}
	return entities
}

// EntityMove 实体移动
// 注意：newX, newY 是格子坐标（不是像素坐标）
func (s *SceneSt) EntityMove(hdl uint64, newX, newY uint32) error {
	e, ok := s.GetEntity(hdl)
	if !ok {
		return fmt.Errorf("entity not found: hdl=%d", hdl)
	}

	// 检查目标位置是否可行走（格子坐标）
	if !s.IsWalkable(int(newX), int(newY)) {
		return fmt.Errorf("position not walkable: (%d, %d)", newX, newY)
	}

	oldPos := e.GetPosition()

	// 更新实体位置
	e.OnMove(newX, newY)

	// 更新AOI
	s.aoiMgr.UpdateEntity(e, oldPos, &argsdef.Position{X: newX, Y: newY})

	return nil
}

// SpawnMonster 生成怪物
// 注意：x, y 是格子坐标（不是像素坐标）
func (s *SceneSt) SpawnMonster(monsterId uint32, x, y uint32) (*entity.MonsterEntity, error) {
	cfgMgr := jsonconf.GetConfigManager()
	monsterCfg, ok := cfgMgr.GetMonsterConfig(monsterId)
	if !ok {
		return nil, fmt.Errorf("monster config not found: %d", monsterId)
	}

	monster := entity.NewMonsterEntity(monsterCfg)
	monster.SetPosition(x, y)

	if err := s.AddEntity(monster); err != nil {
		return nil, err
	}

	s.monsters[monster.GetHdl()] = monster

	log.Infof("Monster spawned: hdl=%d, monsterId=%d, name=%s, pos=(%d,%d)", monster.GetHdl(), monsterId, monsterCfg.Name, x, y)

	return monster, nil
}

// RemoveMonster 移除怪物
func (s *SceneSt) RemoveMonster(hdl uint64) {
	delete(s.monsters, hdl)

	s.RemoveEntity(hdl)
}

// GetMonsterCount 获取怪物数量
func (s *SceneSt) GetMonsterCount() int {
	return len(s.monsters)
}

// InitMonsters 初始化场景怪物（根据MonsterSceneConfig配置）
func (s *SceneSt) InitMonsters() {
	configMgr := jsonconf.GetConfigManager()
	monsterSceneConfigs := configMgr.GetMonsterSceneConfigs(s.sceneId)

	// 查找该场景的怪物配置
	for _, cfg := range monsterSceneConfigs {
		// 根据配置生成怪物
		for i := 0; i < int(cfg.Count); i++ {
			var x, y uint32
			if cfg.BornArea != nil && cfg.BornArea.X2 > cfg.BornArea.X1 && cfg.BornArea.Y2 > cfg.BornArea.Y1 {
				// 从出生点范围随机选择
				x = cfg.BornArea.X1 + uint32(rand.Intn(int(cfg.BornArea.X2-cfg.BornArea.X1)))
				y = cfg.BornArea.Y1 + uint32(rand.Intn(int(cfg.BornArea.Y2-cfg.BornArea.Y1)))
			} else {
				// 使用随机可行走位置
				x, y = s.GetRandomWalkablePos()
			}
			_, err := s.SpawnMonster(cfg.MonsterId, x, y)
			if err != nil {
				log.Warnf("Failed to spawn monster %d in scene %d: %v", cfg.MonsterId, s.sceneId, err)
			}
		}
		log.Infof("Scene %d: Spawned %d monsters (monsterId=%d)", s.sceneId, cfg.Count, cfg.MonsterId)
	}

	// 如果没有找到配置，记录日志
	if s.GetMonsterCount() == 0 {
		log.Infof("Scene %d: No monster configuration found", s.sceneId)
	}
}

// spawnScene2Monsters 生成场景2的怪物
func (s *SceneSt) spawnScene2Monsters() {
	// 史莱姆 x10
	for i := 0; i < 10; i++ {
		x, y := s.GetRandomWalkablePos()
		s.SpawnMonster(10001, x, y)
	}

	// 哥布林 x8
	for i := 0; i < 8; i++ {
		x, y := s.GetRandomWalkablePos()
		s.SpawnMonster(10002, x, y)
	}

	// 森林狼 x5
	for i := 0; i < 5; i++ {
		x, y := s.GetRandomWalkablePos()
		s.SpawnMonster(10003, x, y)
	}

	// 哥布林首领 x2
	for i := 0; i < 2; i++ {
		x, y := s.GetRandomWalkablePos()
		s.SpawnMonster(20001, x, y)
	}

	// 森林守护者 x1 (BOSS)
	x, y := s.GetRandomWalkablePos()
	s.SpawnMonster(30001, x, y)

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

func clampInt(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
