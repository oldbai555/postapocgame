/**
 * @Author: zjj
 * @Date: 2025/11/12
 * @Desc: 怪物场景配置
**/

package jsonconf

// MonsterSceneConfig 怪物场景配置
type MonsterSceneConfig struct {
	MonsterId   uint32    `json:"monsterId"`   // 怪物ID
	SceneId     uint32    `json:"sceneId"`     // 场景ID
	MonsterName string    `json:"monsterName"` // 怪物名称（可选，用于显示）
	BornArea    *BornArea `json:"bornArea"`    // 出生点范围
	Count       uint32    `json:"count"`       // 生成数量
}
