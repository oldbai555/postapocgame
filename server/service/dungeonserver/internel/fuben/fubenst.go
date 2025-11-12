package fuben

import (
	"fmt"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"postapocgame/server/service/dungeonserver/internel/scene"
	"postapocgame/server/service/dungeonserver/internel/scenemgr"
	"sync"
	"time"
)

// FuBenSt 副本结构
type FuBenSt struct {
	fbId   uint32
	name   string
	fbType uint32
	state  uint32

	// 场景管理
	sceneMgr *scenemgr.SceneStMgr

	// 限时副本相关
	createTime  time.Time
	expireTime  time.Time
	maxDuration time.Duration // 最大存在时间

	// 玩家相关
	playerCount int
	maxPlayers  int // 最大玩家数，0表示无限制

	mu sync.RWMutex
}

// NewFuBenSt 创建副本
func NewFuBenSt(fbId uint32, name string, fbType uint32, maxPlayers int, maxDuration time.Duration) *FuBenSt {
	fb := &FuBenSt{
		fbId:        fbId,
		name:        name,
		fbType:      fbType,
		state:       uint32(protocol.FuBenState_FuBenStateNormal),
		sceneMgr:    scenemgr.NewSceneStMgr(),
		createTime:  time.Now(),
		maxPlayers:  maxPlayers,
		maxDuration: maxDuration,
		playerCount: 0,
	}

	// 如果是限时副本，设置过期时间
	if fbType == uint32(protocol.FuBenType_FuBenTypeTimed) || fbType == uint32(protocol.FuBenType_FuBenTypeTimedSingle) || fbType == uint32(protocol.FuBenType_FuBenTypeTimedMulti) {
		if maxDuration > 0 {
			fb.expireTime = fb.createTime.Add(maxDuration)
		}
	}

	return fb
}

// InitScenes 初始化场景
func (fb *FuBenSt) InitScenes(sceneConfigs []jsonconf.SceneConfig) {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	for _, cfg := range sceneConfigs {
		scene := scene.NewSceneSt(cfg.SceneId, fb.fbId, cfg.Name, cfg.WIdth, cfg.Height)
		fb.sceneMgr.AddScene(scene)

		// 初始化场景怪物
		scene.InitMonsters()

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
	fb.mu.RLock()
	defer fb.mu.RUnlock()

	// 检查副本状态
	if fb.state != uint32(protocol.FuBenState_FuBenStateNormal) {
		return false
	}

	// 检查人数限制
	if fb.maxPlayers > 0 && fb.playerCount >= fb.maxPlayers {
		return false
	}

	// 检查是否过期
	if !fb.expireTime.IsZero() && time.Now().After(fb.expireTime) {
		return false
	}

	return true
}

// OnPlayerEnter 玩家进入
func (fb *FuBenSt) OnPlayerEnter() error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.state != uint32(protocol.FuBenState_FuBenStateNormal) {
		return fmt.Errorf("fuben is not available")
	}

	if fb.maxPlayers > 0 && fb.playerCount >= fb.maxPlayers {
		return fmt.Errorf("fuben is full")
	}

	fb.playerCount++
	log.Infof("Player entered FuBen %d, current players: %d", fb.fbId, fb.playerCount)

	return nil
}

// OnPlayerLeave 玩家离开
func (fb *FuBenSt) OnPlayerLeave() {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.playerCount > 0 {
		fb.playerCount--
	}

	log.Infof("Player left FuBen %d, current players: %d", fb.fbId, fb.playerCount)

	// 如果是单人副本且没人了，标记为可关闭
	if fb.fbType == uint32(protocol.FuBenType_FuBenTypeTimedSingle) && fb.playerCount == 0 {
		fb.state = uint32(protocol.FuBenState_FuBenStateClosing)
	}
}

// GetPlayerCount 获取玩家数量
func (fb *FuBenSt) GetPlayerCount() int {
	fb.mu.RLock()
	defer fb.mu.RUnlock()
	return fb.playerCount
}

// IsExpired 检查是否过期
func (fb *FuBenSt) IsExpired() bool {
	fb.mu.RLock()
	defer fb.mu.RUnlock()

	if fb.expireTime.IsZero() {
		return false
	}

	return time.Now().After(fb.expireTime)
}

// Close 关闭副本
func (fb *FuBenSt) Close() {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	fb.state = uint32(protocol.FuBenState_FuBenStateClosed)

	// TODO: 踢出所有玩家
	// TODO: 清理场景数据

	log.Infof("FuBen %d closed", fb.fbId)
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
	fb.mu.RLock()
	defer fb.mu.RUnlock()
	return fb.state
}
