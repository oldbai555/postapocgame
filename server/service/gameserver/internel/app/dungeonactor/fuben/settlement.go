package fuben

import (
	"context"
	"math/rand"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	gshare "postapocgame/server/service/gameserver/internel/core/gshare"
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

	// 发送副本结算消息给 PlayerActor
	protoRewards := make([]*protocol.RewardItem, 0, len(rewards))
	for _, reward := range rewards {
		protoRewards = append(protoRewards, &protocol.RewardItem{
			Type:   reward.Type,
			ItemId: reward.ItemID,
			Count:  reward.Count,
		})
	}
	sendSettleDungeonToPlayerActor(settlement.SessionId, settlement.RoleId, settlement.DungeonID, settlement.Difficulty, settlement.Success, settlement.KillCount, settlement.TimeUsed, protoRewards)

	log.Infof("Dungeon settled: RoleID=%d, DungeonID=%d, Rewards=%d",
		settlement.RoleId, settlement.DungeonID, len(rewards))

	return nil
}

// sendSettleDungeonToPlayerActor 发送副本结算消息给 PlayerActor
func sendSettleDungeonToPlayerActor(sessionId string, roleId uint64, dungeonId uint32, difficulty uint32, success bool, killCount uint32, timeUsed uint32, rewards []*protocol.RewardItem) {
	req := &protocol.D2GSettleDungeonReq{
		SessionId:  sessionId,
		RoleId:     roleId,
		DungeonId:  dungeonId,
		Difficulty: difficulty,
		Success:    success,
		KillCount:  killCount,
		TimeUsed:   timeUsed,
		Rewards:    rewards,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		log.Errorf("[settlement] marshal D2GSettleDungeonReq failed: %v", err)
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, gshare.ContextKeySession, sessionId)
	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdSettleDungeon), data)
	if err := gshare.SendMessageAsync(sessionId, actorMsg); err != nil {
		log.Errorf("[settlement] send SettleDungeon message to PlayerActor failed: sessionId=%s err=%v", sessionId, err)
	}
}
