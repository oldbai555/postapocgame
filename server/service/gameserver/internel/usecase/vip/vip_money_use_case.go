package vip

import (
	"context"

	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	vipdomain "postapocgame/server/service/gameserver/internel/domain/vip"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// VipMoneyUseCaseImpl 实现 MoneyUseCase，用于处理 VIP 经验
type VipMoneyUseCaseImpl struct {
	playerRepo    repository.PlayerRepository
	configManager interfaces.ConfigManager
}

// NewVipMoneyUseCaseImpl 创建 VIP MoneyUseCase 实现
func NewVipMoneyUseCaseImpl(playerRepo repository.PlayerRepository, cfg interfaces.ConfigManager) interfaces.MoneyUseCase {
	return &VipMoneyUseCaseImpl{
		playerRepo:    playerRepo,
		configManager: cfg,
	}
}

// 确保实现接口
var _ interfaces.MoneyUseCase = (*VipMoneyUseCaseImpl)(nil)

// UpdateExp 更新 VIP 经验并处理升级
func (uc *VipMoneyUseCaseImpl) UpdateExp(ctx context.Context, roleID uint64, exp int64) error {
	if exp == 0 {
		return nil
	}
	if exp < 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "vip exp must be positive")
	}

	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return customerr.Wrap(err)
	}

	vipData := vipdomain.EnsureVipData(binaryData)
	if vipData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "vip data not initialized")
	}

	// 添加经验（保证不会溢出到负数）
	vipData.Exp += uint32(exp)

	// 循环检查升级（可能一次获得大量经验）
	for {
		rawCfg, ok := uc.configManager.GetVipConfig(vipData.Level)
		if !ok || rawCfg == nil {
			// 没有更高等级配置，视为已满级
			break
		}
		cfg, ok := rawCfg.(*jsonconf.VipConfig)
		if !ok || cfg == nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "invalid vip config type")
		}

		if uint64(vipData.Exp) < cfg.ExpNeeded {
			break
		}

		// 扣除升级所需经验并升级
		vipData.Exp -= uint32(cfg.ExpNeeded)
		vipData.Level++
	}

	return nil
}

// GetExp 获取当前 VIP 经验
func (uc *VipMoneyUseCaseImpl) GetExp(ctx context.Context, roleID uint64) (int64, error) {
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return 0, customerr.Wrap(err)
	}
	vipData := vipdomain.EnsureVipData(binaryData)
	if vipData == nil {
		return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "vip data not initialized")
	}
	return int64(vipData.Exp), nil
}
