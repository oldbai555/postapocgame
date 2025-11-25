/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package jsonconf

// SceneConfig 场景配置
type SceneConfig struct {
	SceneId  uint32    `json:"sceneId"`  // 场景ID
	Name     string    `json:"name"`     // 场景名称
	Width    int       `json:"width"`    // 场景宽度
	Height   int       `json:"height"`   // 场景高度
	MapId    uint32    `json:"mapId"`    // 关联的地图配置ID
	BornArea *BornArea `json:"bornArea"` // 出生点范围
	GameMap  *GameMap  `json:"-"`        // 运行期挂载的地图数据
}

// BornArea 出生点范围（矩形区域）
type BornArea struct {
	X1 uint32 `json:"x1"` // 左上角X坐标
	Y1 uint32 `json:"y1"` // 左上角Y坐标
	X2 uint32 `json:"x2"` // 右下角X坐标
	Y2 uint32 `json:"y2"` // 右下角Y坐标
}
