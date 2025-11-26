package publicactor

import (
	"context"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/publicactor/offlinedata"
)

const offlineDataFlushIntervalMs int64 = 60 * 1000

func (pr *PublicRole) ensureOfflineDataManager() {
	if pr.offlineDataMgr == nil {
		pr.offlineDataMgr = offlinedata.NewManager()
	}
}

// LoadOfflineData 从数据库加载离线数据
func (pr *PublicRole) LoadOfflineData(ctx context.Context) {
	pr.ensureOfflineDataManager()
	if err := pr.offlineDataMgr.LoadAll(ctx); err != nil {
		log.Warnf("Failed to load offline data: %v", err)
	} else {
		log.Infof("Offline data loaded")
	}
}

func (pr *PublicRole) flushOfflineDataIfNeeded(ctx context.Context, now int64) {
	if pr.offlineDataMgr == nil {
		return
	}
	if pr.lastOfflineDataFlushTime != 0 && now-pr.lastOfflineDataFlushTime < offlineDataFlushIntervalMs {
		return
	}
	pr.offlineDataMgr.FlushDirty(ctx)
	pr.lastOfflineDataFlushTime = now
}

// GetOfflineDataSnapshot 获取指定类型的离线数据
func (pr *PublicRole) GetOfflineDataSnapshot(roleID uint64, dataType protocol.OfflineDataType, target proto.Message) (bool, error) {
	if pr.offlineDataMgr == nil {
		return false, nil
	}
	return pr.offlineDataMgr.GetProto(roleID, dataType, target)
}

func (pr *PublicRole) updateOfflineData(data *protocol.UpdateOfflineDataMsg) {
	if data == nil || data.RoleId == 0 {
		return
	}
	pr.ensureOfflineDataManager()
	if err := pr.offlineDataMgr.UpdateRaw(data.RoleId, data.DataType, data.Payload, data.UpdatedAt, data.Version); err != nil {
		log.Errorf("Failed to update offline data role=%d type=%d: %v", data.RoleId, data.DataType, err)
	}
}

// RegisterOfflineDataHandlers 注册离线数据相关消息
func RegisterOfflineDataHandlers(facade gshare.IPublicActorFacade) {
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdUpdateOfflineData), handleUpdateOfflineDataMsg)
}

var (
	handleUpdateOfflineDataMsg = withPublicRole(handleUpdateOfflineData)
)

func handleUpdateOfflineData(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	updateMsg := &protocol.UpdateOfflineDataMsg{}
	if err := proto.Unmarshal(msg.GetData(), updateMsg); err != nil {
		log.Errorf("Failed to unmarshal UpdateOfflineDataMsg: %v", err)
		return
	}
	publicRole.updateOfflineData(updateMsg)
}
