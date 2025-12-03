package money

import (
	"context"
	"errors"
	"testing"

	"postapocgame/server/internal/protocol"
)

// mockPlayerRepository 模拟 PlayerRepository
type mockPlayerRepository struct {
	binaryData *protocol.PlayerRoleBinaryData
	err        error
}

func (m *mockPlayerRepository) GetBinaryData(ctx context.Context, roleID uint64) (*protocol.PlayerRoleBinaryData, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.binaryData == nil {
		m.binaryData = &protocol.PlayerRoleBinaryData{}
	}
	return m.binaryData, nil
}

func TestInitMoneyDataUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	roleID := uint64(12345)

	t.Run("初始化空的 MoneyData", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{},
		}
		uc := NewInitMoneyDataUseCase(mockRepo)

		err := uc.Execute(ctx, roleID)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		if mockRepo.binaryData.MoneyData == nil {
			t.Fatal("MoneyData should be initialized")
		}
		if mockRepo.binaryData.MoneyData.MoneyMap == nil {
			t.Fatal("MoneyMap should be initialized")
		}
		// 检查默认金币是否注入
		defaultGoldID := uint32(protocol.MoneyType_MoneyTypeGoldCoin)
		if amount, ok := mockRepo.binaryData.MoneyData.MoneyMap[defaultGoldID]; !ok || amount != 100000 {
			t.Errorf("Default gold not injected: got %v, want 100000", amount)
		}
	})

	t.Run("已存在 MoneyData 但 MoneyMap 为空时注入默认金币", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				MoneyData: &protocol.SiMoneyData{
					MoneyMap: make(map[uint32]int64),
				},
			},
		}
		uc := NewInitMoneyDataUseCase(mockRepo)

		err := uc.Execute(ctx, roleID)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		defaultGoldID := uint32(protocol.MoneyType_MoneyTypeGoldCoin)
		if amount, ok := mockRepo.binaryData.MoneyData.MoneyMap[defaultGoldID]; !ok || amount != 100000 {
			t.Errorf("Default gold not injected: got %v, want 100000", amount)
		}
	})

	t.Run("已存在 MoneyData 且 MoneyMap 不为空时不覆盖", func(t *testing.T) {
		existingAmount := int64(50000)
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				MoneyData: &protocol.SiMoneyData{
					MoneyMap: map[uint32]int64{
						uint32(protocol.MoneyType_MoneyTypeGoldCoin): existingAmount,
					},
				},
			},
		}
		uc := NewInitMoneyDataUseCase(mockRepo)

		err := uc.Execute(ctx, roleID)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		defaultGoldID := uint32(protocol.MoneyType_MoneyTypeGoldCoin)
		if amount := mockRepo.binaryData.MoneyData.MoneyMap[defaultGoldID]; amount != existingAmount {
			t.Errorf("Existing gold amount should not be overwritten: got %v, want %v", amount, existingAmount)
		}
	})

	t.Run("Repository 返回错误时应该返回错误", func(t *testing.T) {
		testErr := errors.New("repository error")
		mockRepo := &mockPlayerRepository{
			err: testErr,
		}
		uc := NewInitMoneyDataUseCase(mockRepo)

		err := uc.Execute(ctx, roleID)
		if err == nil {
			t.Fatal("Execute() should return error when repository fails")
		}
		if err != testErr {
			t.Errorf("Execute() error = %v, want %v", err, testErr)
		}
	})
}
