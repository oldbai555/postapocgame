package publicactor

import (
	"context"
	"postapocgame/server/service/gameserver/internel/core/gshare"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
)

// 排行榜相关逻辑

// UpdateRankSnapshot 更新排行榜快照
func (pr *PublicRole) UpdateRankSnapshot(roleId uint64, snapshot *protocol.PlayerRankSnapshot) {
	if snapshot == nil {
		return
	}
	if snapshot.RoleId == 0 {
		snapshot.RoleId = roleId
	}
	pr.ensureOfflineDataManager()
	if err := pr.offlineDataMgr.UpdateProto(roleId, protocol.OfflineDataType_OfflineDataTypeRankSnapshot, snapshot, snapshot.UpdatedAt, 1); err != nil {
		log.Errorf("Failed to update rank snapshot offline data role=%d: %v", roleId, err)
	}
}

// GetRankSnapshot 获取排行榜快照
func (pr *PublicRole) GetRankSnapshot(roleId uint64) (*protocol.PlayerRankSnapshot, bool) {
	if pr.offlineDataMgr == nil {
		return nil, false
	}
	snapshot := &protocol.PlayerRankSnapshot{}
	ok, err := pr.offlineDataMgr.GetProto(roleId, protocol.OfflineDataType_OfflineDataTypeRankSnapshot, snapshot)
	if err != nil {
		log.Warnf("Failed to load rank snapshot role=%d: %v", roleId, err)
		return nil, false
	}
	if !ok {
		return nil, false
	}
	return snapshot, true
}

// UpdateRankValue 更新排行榜数值
func (pr *PublicRole) UpdateRankValue(rankType protocol.RankType, key int64, value int64) {
	rankData, _ := pr.getOrCreateRankData(rankType)

	// 更新或插入排行榜项
	found := false
	for i, item := range rankData.Items {
		if item.Key == key {
			rankData.Items[i].Value = value
			found = true
			break
		}
	}
	if !found {
		rankData.Items = append(rankData.Items, &protocol.RankItem{
			Key:   key,
			Value: value,
		})
	}

	// 排序（降序）
	pr.sortRankData(rankData)
	rankData.UpdatedAt = servertime.UnixMilli()
}

// getOrCreateRankData 获取或创建排行榜数据
func (pr *PublicRole) getOrCreateRankData(rankType protocol.RankType) (*protocol.RankData, bool) {
	value, ok := pr.rankDataMap.Load(rankType)
	if ok {
		return value.(*protocol.RankData), true
	}
	rankData := &protocol.RankData{
		RankType:  rankType,
		Items:     make([]*protocol.RankItem, 0),
		UpdatedAt: servertime.UnixMilli(),
	}
	pr.rankDataMap.Store(rankType, rankData)
	return rankData, false
}

// sortRankData 对排行榜数据进行排序（降序）
func (pr *PublicRole) sortRankData(rankData *protocol.RankData) {
	// 简单的冒泡排序（可以后续优化为更高效的排序算法）
	items := rankData.Items
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].Value < items[j].Value {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

// GetRankData 获取排行榜数据
func (pr *PublicRole) GetRankData(rankType protocol.RankType) *protocol.RankData {
	rankData, _ := pr.getOrCreateRankData(rankType)
	return rankData
}

// ===== 排行榜相关 handler 注册（无闭包捕获 PublicRole） =====

// RegisterRankHandlers 注册排行榜相关的消息处理器
func RegisterRankHandlers(facade gshare.IPublicActorFacade) {
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdUpdateRankSnapshot), handleUpdateRankSnapshotMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdUpdateRankValue), handleUpdateRankValueMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdQueryRank), handleQueryRankMsg)
}

// 排行榜 handler 适配
var (
	handleUpdateRankSnapshotMsg = withPublicRole(handleUpdateRankSnapshot)
	handleUpdateRankValueMsg    = withPublicRole(handleUpdateRankValue)
	handleQueryRankMsg          = withPublicRole(handleQueryRank)
)

// ===== 排行榜业务 handler（从 message_handler.go 迁移）=====

// handleUpdateRankSnapshot 处理更新排行榜快照
func handleUpdateRankSnapshot(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	updateMsg := &protocol.UpdateRankSnapshotMsg{}
	if err := proto.Unmarshal(data, updateMsg); err != nil {
		log.Errorf("Failed to unmarshal UpdateRankSnapshotMsg: %v", err)
		return
	}
	publicRole.UpdateRankSnapshot(updateMsg.RoleId, updateMsg.Snapshot)
}

// handleUpdateRankValue 处理更新排行榜数值
func handleUpdateRankValue(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	updateMsg := &protocol.UpdateRankValueMsg{}
	if err := proto.Unmarshal(data, updateMsg); err != nil {
		log.Errorf("Failed to unmarshal UpdateRankValueMsg: %v", err)
		return
	}
	publicRole.UpdateRankValue(updateMsg.RankType, updateMsg.Key, updateMsg.Value)
}

// handleQueryRank 处理查询排行榜
func handleQueryRank(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	queryMsg := &protocol.QueryRankReqMsg{}
	if err := proto.Unmarshal(data, queryMsg); err != nil {
		log.Errorf("Failed to unmarshal QueryRankReqMsg: %v", err)
		return
	}

	// 获取排行榜数据
	rankData := publicRole.GetRankData(queryMsg.RankType)
	if rankData == nil {
		rankData = &protocol.RankData{
			RankType:  queryMsg.RankType,
			Items:     make([]*protocol.RankItem, 0),
			UpdatedAt: servertime.UnixMilli(),
		}
	}

	// 限制返回数量
	topN := int(queryMsg.TopN)
	if topN <= 0 {
		topN = 100
	}
	if topN > 1000 {
		topN = 1000
	}

	// 获取前N名
	var resultItems []*protocol.RankItem
	if len(rankData.Items) > topN {
		resultItems = rankData.Items[:topN]
	} else {
		resultItems = rankData.Items
	}

	// 组装结果数据（包含快照信息）
	resultRankData := &protocol.RankData{
		RankType:  queryMsg.RankType,
		Items:     resultItems,
		UpdatedAt: rankData.UpdatedAt,
	}

	// 查找请求者的排名和数值
	requesterRank := int64(-1)
	requesterValue := int64(0)
	requesterKey := int64(queryMsg.RequesterRoleId)
	for i, item := range rankData.Items {
		if item.Key == requesterKey {
			requesterRank = int64(i + 1) // 排名从1开始
			requesterValue = item.Value
			break
		}
	}

	// 构建响应消息
	resp := &protocol.S2CQueryRankResultReq{
		RankData:       resultRankData,
		RequesterRank:  requesterRank,
		RequesterValue: requesterValue,
	}
	respData, err := proto.Marshal(resp)
	if err != nil {
		log.Errorf("Failed to marshal S2CQueryRankResultReq: %v", err)
		return
	}

	// 发送给请求者
	if err := sendClientMessageViaPlayerActor(queryMsg.RequesterSessionId, uint16(protocol.S2CProtocol_S2CQueryRankResult), respData); err != nil {
		logSendFailure(queryMsg.RequesterSessionId, uint16(protocol.S2CProtocol_S2CQueryRankResult), err)
		return
	}

	log.Debugf("handleQueryRank: sent rank data for type %d, top %d, requester rank %d", queryMsg.RankType, topN, requesterRank)
}
