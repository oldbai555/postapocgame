package dungeonserverlink

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
)

var dungeonRPC *DungeonClient

func StartDungeonClient(ctx context.Context, dungeonServerAddrMap map[uint8]string) {
	dungeonRPC = NewDungeonClient()
	// 连接到DungeonServer
	for srvType, addr := range dungeonServerAddrMap {
		if err := dungeonRPC.Connect(ctx, srvType, addr); err != nil {
			log.Errorf("connect to dungeon service failed: srvType=%d, addr=%s, err=%v", srvType, addr, err)
		}
	}
}

// Stop 停止DungeonClient
func Stop() {
	if dungeonRPC == nil {
		return
	}
	err := dungeonRPC.Close()
	if err != nil {
		log.Errorf("err:%v", err)
	}
}

func AsyncCall(ctx context.Context, srvType uint8, sessionId string, msgId uint16, data []byte) error {
	if dungeonRPC == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "dungeonRPC not initialized")
	}
	return dungeonRPC.AsyncCall(ctx, srvType, sessionId, msgId, data)
}
