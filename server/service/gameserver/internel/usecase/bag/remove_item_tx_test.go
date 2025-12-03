package bag

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

func TestRemoveItemTxUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	roleID := uint64(12345)

	t.Run("移除单个物品", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				BagData: &protocol.SiBagData{
					Items: []*protocol.ItemSt{
						{ItemId: 1001, Count: 5},
					},
				},
			},
		}
		uc := NewRemoveItemTxUseCase(mockRepo)

		err := uc.Execute(ctx, roleID, 1001, 3)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		if len(mockRepo.binaryData.BagData.Items) != 1 {
			t.Fatalf("Items count = %d, want 1", len(mockRepo.binaryData.BagData.Items))
		}
		if mockRepo.binaryData.BagData.Items[0].Count != 2 {
			t.Errorf("Item Count = %d, want 2", mockRepo.binaryData.BagData.Items[0].Count)
		}
	})

	t.Run("完全移除物品", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				BagData: &protocol.SiBagData{
					Items: []*protocol.ItemSt{
						{ItemId: 1001, Count: 5},
						{ItemId: 1002, Count: 3},
					},
				},
			},
		}
		uc := NewRemoveItemTxUseCase(mockRepo)

		err := uc.Execute(ctx, roleID, 1001, 5)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		if len(mockRepo.binaryData.BagData.Items) != 1 {
			t.Fatalf("Items count = %d, want 1", len(mockRepo.binaryData.BagData.Items))
		}
		if mockRepo.binaryData.BagData.Items[0].ItemId != 1002 {
			t.Errorf("Remaining item ItemId = %d, want 1002", mockRepo.binaryData.BagData.Items[0].ItemId)
		}
	})

	t.Run("跨多个物品移除", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				BagData: &protocol.SiBagData{
					Items: []*protocol.ItemSt{
						{ItemId: 1001, Count: 3},
						{ItemId: 1001, Count: 4},
						{ItemId: 1002, Count: 2},
					},
				},
			},
		}
		uc := NewRemoveItemTxUseCase(mockRepo)

		err := uc.Execute(ctx, roleID, 1001, 5)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		// 应该移除第一个物品（3个），第二个物品剩余2个（4-2=2）
		if len(mockRepo.binaryData.BagData.Items) != 2 {
			t.Fatalf("Items count = %d, want 2", len(mockRepo.binaryData.BagData.Items))
		}
		// 检查剩余物品
		found := false
		for _, item := range mockRepo.binaryData.BagData.Items {
			if item.ItemId == 1001 && item.Count == 2 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Item 1001 with count 2 should exist")
		}
	})

	t.Run("物品不足时返回错误", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				BagData: &protocol.SiBagData{
					Items: []*protocol.ItemSt{
						{ItemId: 1001, Count: 3},
					},
				},
			},
		}
		uc := NewRemoveItemTxUseCase(mockRepo)

		err := uc.Execute(ctx, roleID, 1001, 5)
		if err == nil {
			t.Fatal("Execute() should return error when item not enough")
		}
	})

	t.Run("物品不存在时返回错误", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				BagData: &protocol.SiBagData{
					Items: []*protocol.ItemSt{
						{ItemId: 1001, Count: 3},
					},
				},
			},
		}
		uc := NewRemoveItemTxUseCase(mockRepo)

		err := uc.Execute(ctx, roleID, 9999, 1)
		if err == nil {
			t.Fatal("Execute() should return error when item not found")
		}
	})

	t.Run("BagData 为空时返回错误", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{},
		}
		uc := NewRemoveItemTxUseCase(mockRepo)

		err := uc.Execute(ctx, roleID, 1001, 1)
		if err == nil {
			t.Fatal("Execute() should return error when BagData is nil")
		}
	})

	t.Run("count为0时直接返回", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				BagData: &protocol.SiBagData{
					Items: []*protocol.ItemSt{
						{ItemId: 1001, Count: 5},
					},
				},
			},
		}
		uc := NewRemoveItemTxUseCase(mockRepo)

		err := uc.Execute(ctx, roleID, 1001, 0)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil when count is 0", err)
		}
		// 物品数量应该不变
		if mockRepo.binaryData.BagData.Items[0].Count != 5 {
			t.Errorf("Item Count = %d, want 5 (unchanged)", mockRepo.binaryData.BagData.Items[0].Count)
		}
	})

	t.Run("Repository 返回错误时应该返回错误", func(t *testing.T) {
		testErr := errors.New("repository error")
		mockRepo := &mockPlayerRepository{
			err: testErr,
		}
		uc := NewRemoveItemTxUseCase(mockRepo)

		err := uc.Execute(ctx, roleID, 1001, 1)
		if err == nil {
			t.Fatal("Execute() should return error when repository fails")
		}
	})
}
