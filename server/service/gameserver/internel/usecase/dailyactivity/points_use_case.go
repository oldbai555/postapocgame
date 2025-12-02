package dailyactivity

import (
	"context"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	domain "postapocgame/server/service/gameserver/internel/domain/dailyactivity"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// PointsUseCase 实现 MoneyUseCase，用于处理活跃点（MoneyTypeActivePoint）
type PointsUseCase struct {
	playerRepo     repository.PlayerRepository
	eventPublisher interfaces.EventPublisher
}

func NewPointsUseCase(playerRepo repository.PlayerRepository, eventPublisher interfaces.EventPublisher) interfaces.MoneyUseCase {
	return &PointsUseCase{
		playerRepo:     playerRepo,
		eventPublisher: eventPublisher,
	}
}

// 确保实现接口
var _ interfaces.MoneyUseCase = (*PointsUseCase)(nil)

// UpdateExp 这里的 exp 代表活跃点增量
func (uc *PointsUseCase) UpdateExp(ctx context.Context, roleID uint64, delta int64) error {
	if delta == 0 {
		return nil
	}
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return customerr.Wrap(err)
	}
	data := domain.EnsureData(binaryData)
	if data == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "daily activity data not initialized")
	}

	now := servertime.Now()
	if domain.NeedReset(data, now) {
		domain.ResetForNewDay(data, now)
	}

	newBalance := data.Balance + delta
	if newBalance < 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "active points not enough")
	}
	data.Balance = newBalance
	if delta > 0 {
		data.TodayPoints += uint32(delta)
	}

	// 同步到 MoneyData 中的活跃点货币余额
	if binaryData.MoneyData == nil {
		binaryData.MoneyData = &protocol.SiMoneyData{
			MoneyMap: make(map[uint32]int64),
		}
	}
	if binaryData.MoneyData.MoneyMap == nil {
		binaryData.MoneyData.MoneyMap = make(map[uint32]int64)
	}
	moneyID := uint32(protocol.MoneyType_MoneyTypeActivePoint)
	binaryData.MoneyData.MoneyMap[moneyID] = data.Balance

	// 发布事件
	if uc.eventPublisher != nil {
		uc.eventPublisher.PublishPlayerEvent(ctx, gevent.OnMoneyChange, map[string]interface{}{
			"money_id": moneyID,
			"amount":   data.Balance,
		})
	}
	return nil
}

// GetExp 返回当前活跃点余额
func (uc *PointsUseCase) GetExp(ctx context.Context, roleID uint64) (int64, error) {
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return 0, customerr.Wrap(err)
	}
	data := domain.EnsureData(binaryData)
	if data == nil {
		return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "daily activity data not initialized")
	}
	return data.Balance, nil
}
