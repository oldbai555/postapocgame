/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package gatewaylink

import (
	"postapocgame/server/service/gameserver/internel/iface"
	"sync"
)

var singleSrv *NetworkHandler
var once sync.Once

func GetSession(sessionId string) iface.ISession {
	if singleSrv == nil {
		return nil
	}
	session := singleSrv.GetSession(sessionId)
	if session == nil {
		return nil
	}
	return session
}
