/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package argsdef

// 坐标系统常量
const (
	// TileSize 格子大小（像素）：客户端单个格子为 128×128 像素
	TileSize = 128
	// TileCenterOffset 格子中心偏移：格子中心点距离格子左上角的像素偏移
	TileCenterOffset = 64
)

// Position 位置（格子坐标）
// 服务端坐标 (X, Y) 代表一个格子，玩家始终在格子正中央
// 对应的像素坐标为 (X*TileSize + TileCenterOffset, Y*TileSize + TileCenterOffset)
type Position struct {
	X uint32 // 格子X坐标
	Y uint32 // 格子Y坐标
}

// TileCoordToPixelX 格子坐标转像素坐标X（中心点）
// tileX: 格子X坐标
// 返回: 像素X坐标（格子中心点）
func TileCoordToPixelX(tileX uint32) uint32 {
	return tileX*TileSize + TileCenterOffset
}

// TileCoordToPixelY 格子坐标转像素坐标Y（中心点）
// tileY: 格子Y坐标
// 返回: 像素Y坐标（格子中心点）
func TileCoordToPixelY(tileY uint32) uint32 {
	return tileY*TileSize + TileCenterOffset
}

// TileCoordToPixel 格子坐标转像素坐标（中心点）
// tileX, tileY: 格子坐标
// 返回: 像素坐标（格子中心点）
func TileCoordToPixel(tileX, tileY uint32) (pixelX, pixelY uint32) {
	return TileCoordToPixelX(tileX), TileCoordToPixelY(tileY)
}

// PixelCoordToTileX 像素坐标转格子坐标X
// pixelX: 像素X坐标
// 返回: 格子X坐标
func PixelCoordToTileX(pixelX uint32) uint32 {
	return pixelX / TileSize
}

// PixelCoordToTileY 像素坐标转格子坐标Y
// pixelY: 像素Y坐标
// 返回: 格子Y坐标
func PixelCoordToTileY(pixelY uint32) uint32 {
	return pixelY / TileSize
}

// PixelCoordToTile 像素坐标转格子坐标
// pixelX, pixelY: 像素坐标
// 返回: 格子坐标
func PixelCoordToTile(pixelX, pixelY uint32) (tileX, tileY uint32) {
	return PixelCoordToTileX(pixelX), PixelCoordToTileY(pixelY)
}

// IsSameTile 判断两个位置是否在同一格子
// pos1, pos2: 要比较的两个位置
// 返回: 如果两个位置在同一格子返回true，否则返回false
func IsSameTile(pos1, pos2 *Position) bool {
	if pos1 == nil || pos2 == nil {
		return false
	}
	return pos1.X == pos2.X && pos1.Y == pos2.Y
}
