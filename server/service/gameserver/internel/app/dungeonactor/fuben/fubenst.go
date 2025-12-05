package fuben

import (
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entity"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entitymgr"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/iface"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/scene"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/scenemgr"
	"time"
)

// FuBenSt 副本结构
type FuBenSt struct {
	fbId       uint32
	name       string
	fbType     uint32
	state      uint32
	difficulty uint32 // 难度: 1=普通 2=精英 3=地狱

	// 场景管理
	sceneMgr *scenemgr.SceneStMgr

	// 限时副本相关
	createTime  time.Time
	expireTime  time.Time
	maxDuration time.Duration // 最大存在时间

	// 玩家相关
	playerCount int
	maxPlayers  int // 最大玩家数，0表示无限制

	// 结算相关
	startTime      time.Time       // 开始时间
	killCount      uint32          // 击杀数量
	playerSessions map[string]bool // 玩家Session列表

	// 世界状态
	isNight         bool
	nextCycleUpdate time.Time
}

// NewFuBenSt 创建副本
func NewFuBenSt(fbId uint32, name string, fbType uint32, maxPlayers int, maxDuration time.Duration) *FuBenSt {
	fb := &FuBenSt{
		fbId:            fbId,
		name:            name,
		fbType:          fbType,
		state:           uint32(protocol.FuBenState_FuBenStateNormal),
		difficulty:      1, // 默认普通难度
		sceneMgr:        scenemgr.NewSceneStMgr(),
		createTime:      servertime.Now(),
		maxPlayers:      maxPlayers,
		maxDuration:     maxDuration,
		playerCount:     0,
		startTime:       servertime.Now(),
		killCount:       0,
		playerSessions:  make(map[string]bool),
		nextCycleUpdate: servertime.Now().Add(5 * time.Minute),
	}

	// 如果是限时副本，设置过期时间
	if fbType == uint32(protocol.FuBenType_FuBenTypeTimed) {
		if maxDuration > 0 {
			fb.expireTime = fb.createTime.Add(maxDuration)
		}
	}

	return fb
}

// InitScenes 初始化场景
func (fb *FuBenSt) InitScenes(sceneConfigs []jsonconf.SceneConfig) {
	for _, cfg := range sceneConfigs {
		sc := scene.NewSceneSt(fb, cfg.SceneId, fb.fbId, cfg.Name, cfg.Width, cfg.Height, cfg.GameMap, cfg.BornArea)
		fb.sceneMgr.AddScene(sc)

		// 场景怪物初始化在需要时动态生成

		log.Infof("FuBen %d: Scene %d initialized", fb.fbId, cfg.SceneId)
	}
}

// GetScene 获取场景
func (fb *FuBenSt) GetScene(sceneId uint32) iface.IScene {
	return fb.sceneMgr.GetScene(sceneId)
}

// GetAllScenes 获取所有场景
func (fb *FuBenSt) GetAllScenes() []iface.IScene {
	return fb.sceneMgr.GetAllScenes()
}

// CanEnter 检查是否可以进入副本
func (fb *FuBenSt) CanEnter() bool {
	// 检查副本状态
	if fb.state != uint32(protocol.FuBenState_FuBenStateNormal) {
		return false
	}

	// 检查人数限制
	if fb.maxPlayers > 0 && fb.playerCount >= fb.maxPlayers {
		return false
	}

	// 检查是否过期
	if !fb.expireTime.IsZero() && servertime.Now().After(fb.expireTime) {
		return false
	}

	return true
}

// OnPlayerEnter 玩家进入
func (fb *FuBenSt) OnPlayerEnter(sessionId string) error {
	if fb.state != uint32(protocol.FuBenState_FuBenStateNormal) {
		return customerr.NewError("fuben is not available")
	}

	if fb.maxPlayers > 0 && fb.playerCount >= fb.maxPlayers {
		return customerr.NewError("fuben is full")
	}

	fb.playerCount++
	fb.playerSessions[sessionId] = true

	// 如果是第一个玩家进入，记录开始时间
	if fb.playerCount == 1 {
		fb.startTime = servertime.Now()
	}

	log.Infof("Player entered FuBen %d, current players: %d", fb.fbId, fb.playerCount)

	return nil
}

// OnPlayerLeave 玩家离开
func (fb *FuBenSt) OnPlayerLeave(sessionId string) {
	if fb.playerCount > 0 {
		fb.playerCount--
	}
	delete(fb.playerSessions, sessionId)

	log.Infof("Player left FuBen %d, current players: %d", fb.fbId, fb.playerCount)

	// 限时副本没人时标记为可关闭
	if fb.fbType == uint32(protocol.FuBenType_FuBenTypeTimed) && fb.playerCount == 0 {
		fb.state = uint32(protocol.FuBenState_FuBenStateClosing)
	}
}

// SetDifficulty 设置难度
func (fb *FuBenSt) SetDifficulty(difficulty uint32) {
	fb.difficulty = difficulty
}

// GetDifficulty 获取难度
func (fb *FuBenSt) GetDifficulty() uint32 {
	return fb.difficulty
}

// AddKillCount 增加击杀数
func (fb *FuBenSt) AddKillCount(count uint32) {
	fb.killCount += count
}

// GetKillCount 获取击杀数
func (fb *FuBenSt) GetKillCount() uint32 {
	return fb.killCount
}

// Complete 完成副本（结算）
func (fb *FuBenSt) Complete(success bool) {
	// 使用Closing状态，后续会自动关闭
	fb.state = uint32(protocol.FuBenState_FuBenStateClosing)

	// 计算用时
	timeUsed := uint32(servertime.Now().Sub(fb.startTime).Seconds())

	// 为每个玩家结算
	for sessionId := range fb.playerSessions {
		// 从sessionId获取entity和roleId
		entityMgr := entitymgr.GetEntityMgr()
		et, ok := entityMgr.GetBySession(sessionId)
		if !ok || et == nil {
			log.Warnf("Entity not found for session %s", sessionId)
			continue
		}

		roleId := et.GetId()
		settlement := &DungeonSettlement{
			SessionId:  sessionId,
			RoleId:     roleId,
			DungeonID:  fb.fbId,
			Difficulty: fb.difficulty,
			Success:    success,
			KillCount:  fb.killCount,
			TimeUsed:   timeUsed,
		}

		// 结算奖励
		if err := SettleDungeon(settlement); err != nil {
			log.Errorf("Settle dungeon failed for session %s: %v", sessionId, err)
		}
	}

	log.Infof("FuBen %d completed: success=%v, killCount=%d, timeUsed=%d",
		fb.fbId, success, fb.killCount, timeUsed)
}

// GetPlayerCount 获取玩家数量
func (fb *FuBenSt) GetPlayerCount() int {
	return fb.playerCount
}

// IsExpired 检查是否过期
func (fb *FuBenSt) IsExpired() bool {
	if fb.expireTime.IsZero() {
		return false
	}

	return servertime.Now().After(fb.expireTime)
}

// Close 关闭副本
func (fb *FuBenSt) Close() {
	fb.state = uint32(protocol.FuBenState_FuBenStateClosed)

	// 踢出所有玩家（将玩家移回默认副本）
	entityMgr := entitymgr.GetEntityMgr()
	for sessionId := range fb.playerSessions {
		roleEntity, ok := entityMgr.GetBySession(sessionId)
		if !ok || roleEntity == nil {
			log.Warnf("Entity not found for session %s when closing FuBen", sessionId)
			continue
		}

		// 从当前场景移除实体
		if currentScene, ok := entityMgr.GetSceneByHandle(roleEntity.GetHdl()); ok && currentScene != nil {
			if err := currentScene.RemoveEntity(roleEntity.GetHdl()); err != nil {
				log.Warnf("remove entity from scene failed: %v", err)
			}
		}

		if reviveScene := entity.GetReviveScene(); reviveScene != nil {
			x, y := reviveScene.GetRandomWalkablePos()
			roleEntity.SetPosition(x, y)
			if err := reviveScene.AddEntity(roleEntity); err != nil {
				log.Warnf("Failed to move entity to revive scene: %v", err)
			}
		}

		// 通知玩家副本已关闭
		// 通过GameServer通知客户端（这里可以发送一个协议通知客户端）
		log.Infof("Player %s kicked from FuBen %d", sessionId, fb.fbId)
	}

	// 清理场景数据
	if fb.sceneMgr != nil {
		allScenes := fb.sceneMgr.GetAllScenes()
		for _, sc := range allScenes {
			// 清理场景中的所有实体（除了玩家，玩家已经移走）
			allEntities := sc.GetAllEntities()
			for _, et := range allEntities {
				// 只清理非玩家实体（怪物、掉落物等）
				if et.GetEntityType() != uint32(protocol.EntityType_EtRole) {
					err := sc.RemoveEntity(et.GetHdl())
					if err != nil {
						log.Errorf("err:%v", err)
					}
				}
			}
		}
		// 清空场景管理器
		fb.sceneMgr = scenemgr.NewSceneStMgr()
	}

	// 清空玩家列表
	fb.playerSessions = make(map[string]bool)
	fb.playerCount = 0

	log.Infof("FuBen %d closed, all players kicked and scenes cleared", fb.fbId)
}

// GetFbId 获取副本Id
func (fb *FuBenSt) GetFbId() uint32 {
	return fb.fbId
}

// GetName 获取副本名称
func (fb *FuBenSt) GetName() string {
	return fb.name
}

// GetFbType 获取副本类型
func (fb *FuBenSt) GetFbType() uint32 {
	return fb.fbType
}

// GetState 获取副本状态
func (fb *FuBenSt) GetState() uint32 {
	return fb.state
}

// RunOne 副本常驻逻辑
func (fb *FuBenSt) RunOne(now time.Time) {
	// 限时副本过期检查
	if fb.fbType == uint32(protocol.FuBenType_FuBenTypeTimed) {
		if !fb.expireTime.IsZero() && now.After(fb.expireTime) {
			// 副本已过期，踢出所有玩家
			if fb.state == uint32(protocol.FuBenState_FuBenStateNormal) {
				fb.state = uint32(protocol.FuBenState_FuBenStateClosing)
				log.Infof("FuBen %d expired, kicking all players", fb.fbId)

				// 通知所有玩家副本已过期
				// TODO: 实现踢出玩家的逻辑
			}
		}
	}

	// 世界周期切换
	if fb.nextCycleUpdate.IsZero() {
		fb.nextCycleUpdate = now.Add(5 * time.Minute)
	}
	if now.After(fb.nextCycleUpdate) {
		fb.isNight = !fb.isNight
		fb.nextCycleUpdate = now.Add(5 * time.Minute)
		log.Infof("FuBen %d world cycle switched, night=%v", fb.fbId, fb.isNight)
	}
}
