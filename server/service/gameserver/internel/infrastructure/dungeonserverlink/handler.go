/**
 * @Author: zjj
 * @Date: 2025/11/25
 * @Desc:
**/

package dungeonserverlink

import "postapocgame/server/pkg/log"

var handler *DungeonMessageHandler

func init() {
	handler = NewDungeonMessageHandler()
}

func RegisterRPCHandler(msgId uint16, fn RPCHandler) {
	if handler == nil {
		log.Errorf("dungeonRPC not initialized, cannot register handler for msgId=%d", msgId)
		return
	}
	handler.RegisterRPCHandler(msgId, fn)
}
