/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package fbmgr

import (
	"fmt"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/fuben"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"time"
)

// FuBenMgr 副本管理器
type FuBenMgr struct {
	fubens map[uint32]iface.IFuBen

	// 限时副本
	timedFubens map[string]iface.IFuBen // key: sessionId (单人) 或 teamId (多人)

	nextFbId uint32
	// 定期清理
	lastCleanup time.Time
}

var (
	globalFuBenMgr *FuBenMgr
)

// GetFuBenMgr 获取全局副本管理器
func GetFuBenMgr() *FuBenMgr {
	if globalFuBenMgr == nil {
		globalFuBenMgr = &FuBenMgr{
			fubens:      make(map[uint32]iface.IFuBen),
			timedFubens: make(map[string]iface.IFuBen),
			nextFbId:    1,
		}
	}
	return globalFuBenMgr
}

// CreateDefaultFuBen 创建默认副本
func (m *FuBenMgr) CreateDefaultFuBen() error {
	// 创建 fbId=0 的默认副本
	defaultFuBen := fuben.NewFuBenSt(0, "默认副本", uint32(protocol.FuBenType_FuBenTypePermanent), 0, 0)

	// 初始化场景
	sceneConfigs := []jsonconf.SceneConfig{
		{SceneId: 1, Name: "新手村", Width: 1028, Height: 1028},
		{SceneId: 2, Name: "森林", Width: 1028, Height: 1028},
	}
	defaultFuBen.InitScenes(sceneConfigs)

	// 添加到管理器
	m.AddFuBen(defaultFuBen)
	fuben.SetDefaultFuBen(defaultFuBen)

	log.Infof("Default FuBen (fbId=0) created with 2 scenes")

	return nil
}

// AddFuBen 添加副本
func (m *FuBenMgr) AddFuBen(fb *fuben.FuBenSt) {
	m.fubens[fb.GetFbId()] = fb
	log.Infof("FuBen added: fbId=%d", fb.GetFbId())
}

// RemoveFuBen 移除副本
func (m *FuBenMgr) RemoveFuBen(fbId uint32) {
	if fb, ok := m.fubens[fbId]; ok {
		fb.Close()
		delete(m.fubens, fbId)
		log.Infof("FuBen removed: fbId=%d", fbId)
	}
}

// GetFuBen 获取副本
func (m *FuBenMgr) GetFuBen(fbId uint32) (iface.IFuBen, bool) {
	fb, ok := m.fubens[fbId]
	return fb, ok
}

// CreateTimedFuBenForPlayer 为玩家创建限时副本
func (m *FuBenMgr) CreateTimedFuBenForPlayer(sessionId string, name string, maxDuration time.Duration) (*fuben.FuBenSt, error) {
	// 检查是否已存在
	if _, exists := m.timedFubens[sessionId]; exists {
		return nil, fmt.Errorf("timed fuben already exists for player")
	}

	// 生成副本Id
	fbId := m.nextFbId
	m.nextFbId++

	// 创建限时副本
	fb := fuben.NewFuBenSt(fbId, name, uint32(protocol.FuBenType_FuBenTypeTimed), 1, maxDuration)

	// 注意：场景初始化在进入副本时完成（handleG2DEnterDungeon中），
	// 因为需要根据DungeonId从配置读取场景信息

	m.timedFubens[sessionId] = fb
	m.AddFuBen(fb)

	log.Infof("Timed FuBen created for player %s: fbId=%d", sessionId, fbId)

	return fb, nil
}

// GetTimedFuBenForPlayer 获取玩家的限时副本
func (m *FuBenMgr) GetTimedFuBenForPlayer(sessionId string) (iface.IFuBen, bool) {
	fb, ok := m.timedFubens[sessionId]
	return fb, ok
}

// CleanupExpiredFubens 清理过期副本
func (m *FuBenMgr) CleanupExpiredFubens() {
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

	// 移除过期副本
	for _, fbId := range toRemove {
		m.RemoveFuBen(fbId)
	}

	if len(toRemove) > 0 {
		log.Infof("Cleaned up %d expired fubens", len(toRemove))
	}
}

// RunOne 驱动所有副本的常驻逻辑
func (m *FuBenMgr) RunOne(now time.Time) {
	for _, fb := range m.fubens {
		if fb != nil {
			fb.RunOne(now)
		}
	}

	if m.lastCleanup.IsZero() || now.Sub(m.lastCleanup) >= time.Minute {
		m.CleanupExpiredFubens()
		m.lastCleanup = now
	}
}

// GetAllFubens 获取所有副本
func (m *FuBenMgr) GetAllFubens() []iface.IFuBen {
	fubens := make([]iface.IFuBen, 0, len(m.fubens))
	for _, fb := range m.fubens {
		fubens = append(fubens, fb)
	}
	return fubens
}

func init() {
	fuben.RegisterTimedFuBenProvider(
		func(sessionId string) (iface.IFuBen, bool) {
			return GetFuBenMgr().GetTimedFuBenForPlayer(sessionId)
		},
		func(sessionId string, name string, maxDuration time.Duration) (*fuben.FuBenSt, error) {
			return GetFuBenMgr().CreateTimedFuBenForPlayer(sessionId, name, maxDuration)
		},
	)
}
