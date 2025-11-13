/**
 * @Author: zjj
 * @Date: 2025/11/13
 * @Desc:
**/

package gshare

var (
	platformId uint32
	srvId      uint32
)

func SetPlatformId(id uint32) {
	platformId = id
}
func SetSrvId(id uint32) {
	srvId = id
}
func GetPlatformId() uint32 {
	return platformId
}
func GetSrvId() uint32 {
	return srvId
}
