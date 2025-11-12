/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package fbmgr

import (
	"context"
	"fmt"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/routine"
	"postapocgame/server/service/dungeonserver/internel/fuben"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"sync"
	"time"
)

// FuBenMgr 副本管理器
type FuBenMgr struct {
	fubens map[uint32]iface.IFuBen
	mu     sync.RWMutex

	// 限时副本
	timedFubens map[string]iface.IFuBen // key: sessionId (单人) 或 teamId (多人)
	timedMu     sync.RWMutex

	nextFbId uint32
}

var (
	globalFuBenMgr *FuBenMgr
	fubenOnce      sync.Once
)

// GetFuBenMgr 获取全局副本管理器
func GetFuBenMgr() *FuBenMgr {
	fubenOnce.Do(func() {
		globalFuBenMgr = &FuBenMgr{
			fubens:      make(map[uint32]iface.IFuBen),
			timedFubens: make(map[string]iface.IFuBen),
			nextFbId:    1,
		}
	})
	return globalFuBenMgr
}

// CreateDefaultFuBen 创建默认副本
func (m *FuBenMgr) CreateDefaultFuBen() error {
	// 创建 fbId=0 的默认副本
	defaultFuBen := fuben.NewFuBenSt(0, "默认副本", uint32(protocol.FuBenType_FuBenTypePermanent), 0, 0)

	// 初始化场景
	sceneConfigs := []jsonconf.SceneConfig{
		{SceneId: 1, Name: "新手村", WIdth: 1028, Height: 1028},
		{SceneId: 2, Name: "森林", WIdth: 1028, Height: 1028},
	}
	defaultFuBen.InitScenes(sceneConfigs)

	// 添加到管理器
	m.AddFuBen(defaultFuBen)

	log.Infof("Default FuBen (fbId=0) created with 2 scenes")

	return nil
}

// AddFuBen 添加副本
func (m *FuBenMgr) AddFuBen(fb *fuben.FuBenSt) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fubens[fb.GetFbId()] = fb
	log.Infof("FuBen added: fbId=%d", fb.GetFbId())
}

// RemoveFuBen 移除副本
func (m *FuBenMgr) RemoveFuBen(fbId uint32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if fb, ok := m.fubens[fbId]; ok {
		fb.Close()
		delete(m.fubens, fbId)
		log.Infof("FuBen removed: fbId=%d", fbId)
	}
}

// GetFuBen 获取副本
func (m *FuBenMgr) GetFuBen(fbId uint32) (iface.IFuBen, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fb, ok := m.fubens[fbId]
	return fb, ok
}

// CreateTimedFuBenForPlayer 为玩家创建限时副本
func (m *FuBenMgr) CreateTimedFuBenForPlayer(sessionId string, name string, maxDuration time.Duration) (*fuben.FuBenSt, error) {
	m.timedMu.Lock()
	defer m.timedMu.Unlock()

	// 检查是否已存在
	if _, exists := m.timedFubens[sessionId]; exists {
		return nil, fmt.Errorf("timed fuben already exists for player")
	}

	// 生成副本Id
	m.mu.Lock()
	fbId := m.nextFbId
	m.nextFbId++
	m.mu.Unlock()

	// 创建限时单人副本
	fb := fuben.NewFuBenSt(fbId, name, uint32(protocol.FuBenType_FuBenTypeTimedSingle), 1, maxDuration)

	// TODO: 根据配置初始化场景

	m.timedFubens[sessionId] = fb
	m.AddFuBen(fb)

	log.Infof("Timed FuBen created for player %s: fbId=%d", sessionId, fbId)

	return fb, nil
}

// GetTimedFuBenForPlayer 获取玩家的限时副本
func (m *FuBenMgr) GetTimedFuBenForPlayer(sessionId string) (iface.IFuBen, bool) {
	m.timedMu.RLock()
	defer m.timedMu.RUnlock()

	fb, ok := m.timedFubens[sessionId]
	return fb, ok
}

// CleanupExpiredFubens 清理过期副本
func (m *FuBenMgr) CleanupExpiredFubens() {
	m.mu.Lock()
	toRemove := make([]uint32, 0)

	for fbId, fb := range m.fubens {
		if fb.GetFbType() != uint32(protocol.FuBenType_FuBenTypePermanent) {
			if fb.IsExpired() || fb.GetState() == uint32(protocol.FuBenState_FuBenStateClosing) {
				if fb.GetPlayerCount() == 0 {
					toRemove = append(toRemove, fbId)
				}
			}
		}
	}
	m.mu.Unlock()

	// 移除过期副本
	for _, fbId := range toRemove {
		m.RemoveFuBen(fbId)
	}

	if len(toRemove) > 0 {
		log.Infof("Cleaned up %d expired fubens", len(toRemove))
	}
}

// StartCleanupRoutine 启动清理协程
func (m *FuBenMgr) StartCleanupRoutine(ctx context.Context) {
	routine.GoV2(func() error {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				break
			case <-ticker.C:
				m.CleanupExpiredFubens()
			}
		}
	})
	log.Infof("FuBen cleanup routine started")
}

// GetAllFubens 获取所有副本
func (m *FuBenMgr) GetAllFubens() []iface.IFuBen {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fubens := make([]iface.IFuBen, 0, len(m.fubens))
	for _, fb := range m.fubens {
		fubens = append(fubens, fb)
	}
	return fubens
}
