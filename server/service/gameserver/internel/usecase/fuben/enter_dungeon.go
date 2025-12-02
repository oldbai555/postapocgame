package fuben

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
	"time"
)

// EnterDungeonUseCase 进入副本用例
type EnterDungeonUseCase struct {
	playerRepo     repository.PlayerRepository
	configManager  interfaces.ConfigManager
	consumeUseCase interfaces.ConsumeUseCase
	dungeonGateway interfaces.DungeonServerGateway
}

// NewEnterDungeonUseCase 创建进入副本用例
func NewEnterDungeonUseCase(
	playerRepo repository.PlayerRepository,
	configManager interfaces.ConfigManager,
	dungeonGateway interfaces.DungeonServerGateway,
) *EnterDungeonUseCase {
	return &EnterDungeonUseCase{
		playerRepo:     playerRepo,
		configManager:  configManager,
		dungeonGateway: dungeonGateway,
	}
}

// SetDependencies 设置依赖（用于注入 ConsumeUseCase）
func (uc *EnterDungeonUseCase) SetDependencies(consumeUseCase interfaces.ConsumeUseCase) {
	uc.consumeUseCase = consumeUseCase
}

// Execute 执行进入副本用例
func (uc *EnterDungeonUseCase) Execute(ctx context.Context, roleID uint64, dungeonID uint32, difficulty uint32) error {
	// 获取副本配置
	dungeonCfgRaw, ok := uc.configManager.GetDungeonConfig(dungeonID)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "副本不存在")
	}

	dungeonCfg, ok := dungeonCfgRaw.(*jsonconf.DungeonConfig)
	if !ok || dungeonCfg == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "invalid dungeon config type")
	}

	// 检查是否为限时副本
	if dungeonCfg.Type != 2 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "该副本不是限时副本")
	}

	// 检查难度是否有效
	var difficultyCfg *jsonconf.DungeonDifficulty
	for i := range dungeonCfg.Difficulties {
		if dungeonCfg.Difficulties[i].Difficulty == difficulty {
			difficultyCfg = dungeonCfg.Difficulties[i]
			break
		}
	}
	if difficultyCfg == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "难度不存在")
	}

	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	// 检查副本系统是否初始化
	if binaryData.DungeonData == nil {
		binaryData.DungeonData = &protocol.SiDungeonData{
			Records: make([]*protocol.DungeonRecord, 0),
		}
	}
	if binaryData.DungeonData.Records == nil {
		binaryData.DungeonData.Records = make([]*protocol.DungeonRecord, 0)
	}

	// 检查每日进入次数
	record := uc.getDungeonRecord(binaryData.DungeonData, dungeonID, difficulty)
	if record != nil {
		now := servertime.Now()
		lastResetTime := time.Unix(record.ResetTime, 0)

		// 检查是否需要重置（每日重置）
		if now.Sub(lastResetTime) >= 24*time.Hour {
			record.EnterCount = 0
			record.ResetTime = now.Unix()
		}

		// 检查每日最大进入次数
		if dungeonCfg.MaxEnterPerDay > 0 && record.EnterCount >= dungeonCfg.MaxEnterPerDay {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "今日进入次数已用完")
		}
	}

	// 检查消耗物品（如通天令）
	if len(difficultyCfg.ConsumeItems) > 0 {
		if uc.consumeUseCase != nil {
			// 检查消耗是否足够
			if err := uc.consumeUseCase.CheckConsume(ctx, roleID, difficultyCfg.ConsumeItems); err != nil {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "消耗物品不足: %v", err)
			}
			// 扣除消耗
			if err := uc.consumeUseCase.ApplyConsume(ctx, roleID, difficultyCfg.ConsumeItems); err != nil {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "扣除消耗失败: %v", err)
			}
		} else {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "consume use case not initialized")
		}
	}

	// 更新进入记录
	if err := uc.updateEnterRecord(binaryData.DungeonData, dungeonID, difficulty); err != nil {
		return err
	}

	log.Infof("Enter dungeon: RoleID=%d, DungeonID=%d, Difficulty=%d", roleID, dungeonID, difficulty)
	return nil
}

// getDungeonRecord 获取副本记录
func (uc *EnterDungeonUseCase) getDungeonRecord(dungeonData *protocol.SiDungeonData, dungeonID uint32, difficulty uint32) *protocol.DungeonRecord {
	if dungeonData == nil || dungeonData.Records == nil {
		return nil
	}
	for _, record := range dungeonData.Records {
		if record != nil && record.DungeonId == dungeonID && record.Difficulty == difficulty {
			return record
		}
	}
	return nil
}

// updateEnterRecord 更新进入记录
func (uc *EnterDungeonUseCase) updateEnterRecord(dungeonData *protocol.SiDungeonData, dungeonID uint32, difficulty uint32) error {
	if dungeonData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "dungeon data not initialized")
	}

	record := uc.getOrCreateDungeonRecord(dungeonData, dungeonID, difficulty)
	if record == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "failed to create dungeon record")
	}

	now := servertime.Now()
	lastResetTime := time.Unix(record.ResetTime, 0)

	// 检查是否需要重置（每日重置）
	if now.Sub(lastResetTime) >= 24*time.Hour {
		record.EnterCount = 1
		record.ResetTime = now.Unix()
	} else {
		record.EnterCount++
	}
	record.LastEnterTime = now.Unix()

	return nil
}

// getOrCreateDungeonRecord 获取或创建副本记录
func (uc *EnterDungeonUseCase) getOrCreateDungeonRecord(dungeonData *protocol.SiDungeonData, dungeonID uint32, difficulty uint32) *protocol.DungeonRecord {
	if dungeonData == nil {
		return nil
	}
	if dungeonData.Records == nil {
		dungeonData.Records = make([]*protocol.DungeonRecord, 0)
	}

	// 查找现有记录
	for _, record := range dungeonData.Records {
		if record != nil && record.DungeonId == dungeonID && record.Difficulty == difficulty {
			return record
		}
	}

	// 创建新记录
	now := servertime.Now().Unix()
	newRecord := &protocol.DungeonRecord{
		DungeonId:     dungeonID,
		Difficulty:    difficulty,
		LastEnterTime: now,
		EnterCount:    0,
		ResetTime:     now,
	}
	dungeonData.Records = append(dungeonData.Records, newRecord)
	return newRecord
}
