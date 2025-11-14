package fuben

import (
	"context"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/gameserverlink"
)

// DungeonSettlement 副本结算
type DungeonSettlement struct {
	SessionId  string // 会话ID
	RoleId     uint64 // 角色ID
	DungeonID  uint32
	Difficulty uint32
	Success    bool   // 是否成功
	KillCount  uint32 // 击杀数量
	TimeUsed   uint32 // 用时（秒）
}

// CalculateRewards 计算副本奖励
func CalculateRewards(settlement *DungeonSettlement) ([]RewardItem, error) {
	dungeonConfig, ok := jsonconf.GetConfigManager().GetDungeonConfig(settlement.DungeonID)
	if !ok {
		log.Warnf("Dungeon config not found: %d", settlement.DungeonID)
		return nil, nil
	}

	// 查找对应难度的配置
	var difficultyConfig *jsonconf.DungeonDifficulty
	for _, diff := range dungeonConfig.Difficulties {
		if diff != nil && diff.Difficulty == settlement.Difficulty {
			difficultyConfig = diff
			break
		}
	}

	if difficultyConfig == nil {
		log.Warnf("Difficulty config not found: dungeon=%d, difficulty=%d", settlement.DungeonID, settlement.Difficulty)
		return nil, nil
	}

	rewards := make([]RewardItem, 0)

	// 计算奖励
	for _, rewardConfig := range difficultyConfig.Rewards {
		// 检查概率
		if rewardConfig.Rate < 1.0 && rand.Float32() > rewardConfig.Rate {
			continue
		}

		switch rewardConfig.Type {
		case 1: // 经验奖励
			// 经验奖励在结算时直接发放，这里只记录
			rewards = append(rewards, RewardItem{
				Type:   1,
				ItemID: 0,
				Count:  rewardConfig.Count,
			})
		case 2: // 金币奖励
			rewards = append(rewards, RewardItem{
				Type:   2,
				ItemID: 0,
				Count:  rewardConfig.Count,
			})
		case 3: // 物品奖励
			rewards = append(rewards, RewardItem{
				Type:   3,
				ItemID: rewardConfig.ItemID,
				Count:  rewardConfig.Count,
			})
		}
	}

	return rewards, nil
}

// RewardItem 奖励物品
type RewardItem struct {
	Type   uint32 // 1=经验 2=金币 3=物品
	ItemID uint32
	Count  uint32
}

// SettleDungeon 结算副本（发放奖励）
func SettleDungeon(settlement *DungeonSettlement) error {
	if !settlement.Success {
		log.Infof("Dungeon failed, no rewards: RoleID=%d, DungeonID=%d", settlement.RoleId, settlement.DungeonID)
		return nil
	}

	// 计算奖励
	rewards, err := CalculateRewards(settlement)
	if err != nil {
		return err
	}

	// 构建RPC请求
	req := &protocol.D2GSettleDungeonReq{
		SessionId:  settlement.SessionId,
		RoleId:     settlement.RoleId,
		DungeonId:  settlement.DungeonID,
		Difficulty: settlement.Difficulty,
		Success:    settlement.Success,
		KillCount:  settlement.KillCount,
		TimeUsed:   settlement.TimeUsed,
		Rewards:    make([]*protocol.RewardItem, 0, len(rewards)),
	}

	// 转换奖励格式
	for _, reward := range rewards {
		req.Rewards = append(req.Rewards, &protocol.RewardItem{
			Type:   reward.Type,
			ItemId: reward.ItemID,
			Count:  reward.Count,
		})
	}

	// 序列化请求
	data, err := proto.Marshal(req)
	if err != nil {
		log.Errorf("Marshal D2GSettleDungeonReq failed: %v", err)
		return err
	}

	// 通过RPC调用GameServer
	ctx := context.Background()
	msgId := uint16(protocol.D2GRpcProtocol_D2GSettleDungeon)
	if err := gameserverlink.CallGameServer(ctx, settlement.SessionId, msgId, data); err != nil {
		log.Errorf("CallGameServer failed: %v", err)
		return err
	}

	log.Infof("Dungeon settled: RoleID=%d, DungeonID=%d, Rewards=%d",
		settlement.RoleId, settlement.DungeonID, len(rewards))

	return nil
}
