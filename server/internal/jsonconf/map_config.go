package jsonconf

import (
	"errors"
	"fmt"
	"math/rand"
)

// MapConfig 地图配置，提供 GameMap 数据源
type MapConfig struct {
	MapId    uint32    `json:"mapId"`    // 地图唯一ID
	TileData *TileData `json:"tileData"` // 地图瓦片数据

	gameMap *GameMap // 运行期缓存的地图
}

// TileData 原始瓦片数据
type TileData struct {
	Row   int     `json:"row"`   // 行数(高度)
	Col   int     `json:"col"`   // 列数(宽度)
	Tiles []uint8 `json:"tiles"` // 瓦片数据（一维展开，0=不可行走，1=可行走）
}

// GameMap 游戏地图数据
type GameMap struct {
	width_           int32
	height_          int32
	tile_            []uint8
	movableIndexList []int32
}

func newGameMapFromTileData(tileData *TileData) (*GameMap, error) {
	if tileData == nil {
		return nil, errors.New("tileData is nil")
	}
	if tileData.Row <= 0 || tileData.Col <= 0 {
		return nil, fmt.Errorf("invalid tile dimensions row=%d col=%d", tileData.Row, tileData.Col)
	}
	expected := tileData.Row * tileData.Col
	if len(tileData.Tiles) != expected {
		return nil, fmt.Errorf("tile count mismatch, expect %d got %d", expected, len(tileData.Tiles))
	}

	gameMap := &GameMap{
		width_:  int32(tileData.Col),
		height_: int32(tileData.Row),
		tile_:   make([]uint8, len(tileData.Tiles)),
	}
	copy(gameMap.tile_, tileData.Tiles)

	gameMap.movableIndexList = make([]int32, 0, len(tileData.Tiles)/2)
	for idx, v := range tileData.Tiles {
		if v == 1 {
			gameMap.movableIndexList = append(gameMap.movableIndexList, int32(idx))
		}
	}

	return gameMap, nil
}

// Width 地图宽度
func (gm *GameMap) Width() int32 {
	return gm.width_
}

// Height 地图高度
func (gm *GameMap) Height() int32 {
	return gm.height_
}

// MovableCount 返回可行走格子数量
func (gm *GameMap) MovableCount() int {
	if gm == nil {
		return 0
	}
	return len(gm.movableIndexList)
}

// IsWalkable 判断坐标是否可行走
func (gm *GameMap) IsWalkable(x, y int32) bool {
	if gm == nil {
		return false
	}
	if x < 0 || x >= gm.width_ || y < 0 || y >= gm.height_ {
		return false
	}
	index := y*gm.width_ + x
	if int(index) >= len(gm.tile_) || index < 0 {
		return false
	}
	return gm.tile_[index] == 1
}

// RandomWalkableCoord 随机返回一个可行走坐标
func (gm *GameMap) RandomWalkableCoord() (int32, int32, bool) {
	if gm == nil || len(gm.movableIndexList) == 0 {
		return 0, 0, false
	}
	idx := gm.movableIndexList[rand.Intn(len(gm.movableIndexList))]
	x := idx % gm.width_
	y := idx / gm.width_
	return x, y, true
}

// ClampCoord 返回裁剪到地图范围内的坐标
func (gm *GameMap) ClampCoord(x, y int64) (int32, int32) {
	if gm == nil {
		return int32(x), int32(y)
	}
	clamp := func(v int64, max int32) int32 {
		if v < 0 {
			return 0
		}
		if v > int64(max) {
			return max
		}
		return int32(v)
	}
	return clamp(x, gm.width_-1), clamp(y, gm.height_-1)
}

// CoordToIndex 返回坐标对应的一维下标（row-major）
func (gm *GameMap) CoordToIndex(x, y int32) (int32, bool) {
	if gm == nil {
		return 0, false
	}
	if x < 0 || x >= gm.width_ || y < 0 || y >= gm.height_ {
		return 0, false
	}
	index := y*gm.width_ + x
	if index < 0 || index >= int32(len(gm.tile_)) {
		return 0, false
	}
	return index, true
}

// IndexToCoord 根据一维下标返回坐标
func (gm *GameMap) IndexToCoord(idx int32) (int32, int32, bool) {
	if gm == nil {
		return 0, 0, false
	}
	if idx < 0 || idx >= int32(len(gm.tile_)) {
		return 0, 0, false
	}
	x := idx % gm.width_
	y := idx / gm.width_
	return x, y, true
}
