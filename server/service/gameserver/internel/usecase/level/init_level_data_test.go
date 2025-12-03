package level

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

func TestInitLevelDataUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	roleID := uint64(12345)

	t.Run("初始化空的 LevelData", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{},
		}
		uc := NewInitLevelDataUseCase(mockRepo)

		err := uc.Execute(ctx, roleID)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		if mockRepo.binaryData.LevelData == nil {
			t.Fatal("LevelData should be initialized")
		}
		if mockRepo.binaryData.LevelData.Level != 1 {
			t.Errorf("Level should be 1, got %d", mockRepo.binaryData.LevelData.Level)
		}
		if mockRepo.binaryData.LevelData.Exp != 0 {
			t.Errorf("Exp should be 0, got %d", mockRepo.binaryData.LevelData.Exp)
		}
	})

	t.Run("修正小于1的等级", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				LevelData: &protocol.SiLevelData{
					Level: 0,
					Exp:   100,
				},
			},
		}
		uc := NewInitLevelDataUseCase(mockRepo)

		err := uc.Execute(ctx, roleID)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		if mockRepo.binaryData.LevelData.Level != 1 {
			t.Errorf("Level should be corrected to 1, got %d", mockRepo.binaryData.LevelData.Level)
		}
	})

	t.Run("修正小于0的经验", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				LevelData: &protocol.SiLevelData{
					Level: 10,
					Exp:   -100,
				},
			},
		}
		uc := NewInitLevelDataUseCase(mockRepo)

		err := uc.Execute(ctx, roleID)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		if mockRepo.binaryData.LevelData.Exp != 0 {
			t.Errorf("Exp should be corrected to 0, got %d", mockRepo.binaryData.LevelData.Exp)
		}
	})

	t.Run("同步经验到货币系统", func(t *testing.T) {
		expValue := int64(5000)
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				LevelData: &protocol.SiLevelData{
					Level: 10,
					Exp:   expValue,
				},
				MoneyData: &protocol.SiMoneyData{
					MoneyMap: make(map[uint32]int64),
				},
			},
		}
		uc := NewInitLevelDataUseCase(mockRepo)

		err := uc.Execute(ctx, roleID)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
		if amount, ok := mockRepo.binaryData.MoneyData.MoneyMap[expMoneyID]; !ok || amount != expValue {
			t.Errorf("Exp not synced to money system: got %v, want %d", amount, expValue)
		}
	})

	t.Run("MoneyData 不存在时不创建", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				LevelData: &protocol.SiLevelData{
					Level: 10,
					Exp:   5000,
				},
			},
		}
		uc := NewInitLevelDataUseCase(mockRepo)

		err := uc.Execute(ctx, roleID)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		if mockRepo.binaryData.MoneyData != nil {
			t.Error("MoneyData should not be created if it doesn't exist")
		}
	})

	t.Run("Repository 返回错误时应该返回错误", func(t *testing.T) {
		testErr := errors.New("repository error")
		mockRepo := &mockPlayerRepository{
			err: testErr,
		}
		uc := NewInitLevelDataUseCase(mockRepo)

		err := uc.Execute(ctx, roleID)
		if err == nil {
			t.Fatal("Execute() should return error when repository fails")
		}
		if err != testErr {
			t.Errorf("Execute() error = %v, want %v", err, testErr)
		}
	})
}
