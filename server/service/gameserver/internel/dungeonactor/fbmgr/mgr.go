/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package fbmgr

import (
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	fuben2 "postapocgame/server/service/gameserver/internel/dungeonactor/fuben"
	"postapocgame/server/service/gameserver/internel/dungeonactor/iface"
	"time"
)

// FuBenMgr 副本管理器
type FuBenMgr struct {
	fubens map[uint32]iface.IFuBen

	// 未来如需扩展多实例副本，可重新引入计数器与定期清理。
}

var (
	globalFuBenMgr *FuBenMgr
)

// GetFuBenMgr 获取全局副本管理器
func GetFuBenMgr() *FuBenMgr {
	if globalFuBenMgr == nil {
		globalFuBenMgr = &FuBenMgr{
			fubens: make(map[uint32]iface.IFuBen),
		}
	}
	return globalFuBenMgr
}

// CreateDefaultFuBen 创建默认副本
func (m *FuBenMgr) CreateDefaultFuBen() error {
	// 创建 fbId=0 的默认副本
	defaultFuBen := fuben2.NewFuBenSt(0, "默认副本", uint32(protocol.FuBenType_FuBenTypePermanent), 0, 0)

	// 初始化场景
	defaultSceneIds := []uint32{1, 2}
	sceneConfigs := make([]jsonconf.SceneConfig, 0, len(defaultSceneIds))
	configMgr := jsonconf.GetConfigManager()
	for _, sceneId := range defaultSceneIds {
		if cfg := configMgr.GetSceneConfig(sceneId); cfg != nil {
			if cfg.GameMap == nil {
				log.Warnf("scene %d missing GameMap, fallback to random map", sceneId)
			}
			sceneConfigs = append(sceneConfigs, *cfg)
		}
	}
	if len(sceneConfigs) == 0 {
		log.Warnf("scene configs not found for default fuben, using fallback definitions")
		sceneConfigs = []jsonconf.SceneConfig{
			{SceneId: 1, Name: "默认场景", Width: 1028, Height: 1028},
		}
	}
	defaultFuBen.InitScenes(sceneConfigs)

	// 添加到管理器
	m.AddFuBen(defaultFuBen)
	fuben2.SetDefaultFuBen(defaultFuBen)

	log.Infof("Default FuBen (fbId=0) created with 2 scenes")

	return nil
}

// AddFuBen 添加副本
func (m *FuBenMgr) AddFuBen(fb *fuben2.FuBenSt) {
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

// RunOne 驱动所有副本的常驻逻辑
func (m *FuBenMgr) RunOne(now time.Time) {
	for _, fb := range m.fubens {
		if fb != nil {
			fb.RunOne(now)
		}
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
