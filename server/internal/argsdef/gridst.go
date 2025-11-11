/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package argsdef

type GrIdSt struct {
	X int
	Y int
}

const GrIdSize = 100 // 每个格子的大小

func GetGrIdSt(pos *Position) GrIdSt {
	return GrIdSt{
		X: int(pos.X) / GrIdSize,
		Y: int(pos.Y) / GrIdSize,
	}
}

// GetNineGrIds 获取九宫格
func GetNineGrIds(pos *Position) []GrIdSt {
	center := GetGrIdSt(pos)

	grIds := make([]GrIdSt, 0, 9)
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			grIds = append(grIds, GrIdSt{
				X: center.X + dx,
				Y: center.Y + dy,
			})
		}
	}

	return grIds
}
