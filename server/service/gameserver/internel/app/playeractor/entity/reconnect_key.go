/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package entity

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"postapocgame/server/internal/servertime"
)

func generateReconnectKey(sessionID string, roleID uint64) string {
	str := fmt.Sprintf("%s_%d_%d", sessionID, roleID, servertime.Now().Unix())
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}
