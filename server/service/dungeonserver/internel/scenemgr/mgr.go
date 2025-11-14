/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package scenemgr

import (
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

// SceneStMgr 场景管理器
type SceneStMgr struct {
	scenes map[uint32]iface.IScene
}

// NewSceneStMgr 创建场景管理器
func NewSceneStMgr() *SceneStMgr {
	return &SceneStMgr{
		scenes: make(map[uint32]iface.IScene),
	}
}

// AddScene 添加场景
func (m *SceneStMgr) AddScene(scene iface.IScene) {
	m.scenes[scene.GetSceneId()] = scene
	log.Infof("Scene added: sceneId=%d", scene.GetSceneId())
}

// RemoveScene 移除场景
func (m *SceneStMgr) RemoveScene(sceneId uint32) {
	delete(m.scenes, sceneId)
	log.Infof("Scene removed: sceneId=%d", sceneId)
}

// GetScene 获取场景
func (m *SceneStMgr) GetScene(sceneId uint32) iface.IScene {
	scene, ok := m.scenes[sceneId]
	if !ok {
		return nil
	}
	return scene
}

// GetAllScenes 获取所有场景
func (m *SceneStMgr) GetAllScenes() []iface.IScene {
	scenes := make([]iface.IScene, 0, len(m.scenes))
	for _, scene := range m.scenes {
		scenes = append(scenes, scene)
	}
	return scenes
}
