package bag

import (
	"context"
	"errors"
	"testing"

	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
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

// mockConfigManager 模拟 ConfigManager
type mockConfigManager struct {
	itemConfig map[uint32]interface{}
	bagConfig  map[uint32]interface{}
}

func (m *mockConfigManager) GetItemConfig(itemID uint32) (interface{}, bool) {
	if m.itemConfig == nil {
		return nil, false
	}
	config, ok := m.itemConfig[itemID]
	return config, ok
}

func (m *mockConfigManager) GetBagConfig(bagType uint32) (interface{}, bool) {
	if m.bagConfig == nil {
		return nil, false
	}
	config, ok := m.bagConfig[bagType]
	return config, ok
}

func TestAddItemTxUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	roleID := uint64(12345)

	t.Run("添加新物品到空背包", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{},
		}
		mockConfig := &mockConfigManager{
			itemConfig: map[uint32]interface{}{
				1001: &jsonconf.ItemConfig{
					ItemId:   1001,
					MaxStack: 1,
				},
			},
			bagConfig: map[uint32]interface{}{
				1: &jsonconf.BagConfig{
					Size: 100,
				},
			},
		}
		uc := NewAddItemTxUseCase(mockRepo, mockConfig)

		err := uc.Execute(ctx, roleID, 1001, 5, 0)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		if mockRepo.binaryData.BagData == nil {
			t.Fatal("BagData should be initialized")
		}
		if len(mockRepo.binaryData.BagData.Items) != 1 {
			t.Fatalf("Items count = %d, want 1", len(mockRepo.binaryData.BagData.Items))
		}
		item := mockRepo.binaryData.BagData.Items[0]
		if item.ItemId != 1001 || item.Count != 5 {
			t.Errorf("Item = %+v, want ItemId=1001 Count=5", item)
		}
	})

	t.Run("堆叠到已有物品", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				BagData: &protocol.SiBagData{
					Items: []*protocol.ItemSt{
						{ItemId: 1001, Count: 3, Bind: 0},
					},
				},
			},
		}
		mockConfig := &mockConfigManager{
			itemConfig: map[uint32]interface{}{
				1001: &jsonconf.ItemConfig{
					ItemId:   1001,
					MaxStack: 10,
				},
			},
		}
		uc := NewAddItemTxUseCase(mockRepo, mockConfig)

		err := uc.Execute(ctx, roleID, 1001, 5, 0)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		if len(mockRepo.binaryData.BagData.Items) != 1 {
			t.Fatalf("Items count = %d, want 1 (stacked)", len(mockRepo.binaryData.BagData.Items))
		}
		item := mockRepo.binaryData.BagData.Items[0]
		if item.Count != 8 {
			t.Errorf("Item Count = %d, want 8 (3+5)", item.Count)
		}
	})

	t.Run("堆叠超过最大堆叠数时创建新物品", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				BagData: &protocol.SiBagData{
					Items: []*protocol.ItemSt{
						{ItemId: 1001, Count: 9, Bind: 0}, // MaxStack=10
					},
				},
			},
		}
		mockConfig := &mockConfigManager{
			itemConfig: map[uint32]interface{}{
				1001: &jsonconf.ItemConfig{
					ItemId:   1001,
					MaxStack: 10,
				},
			},
			bagConfig: map[uint32]interface{}{
				1: &jsonconf.BagConfig{Size: 100},
			},
		}
		uc := NewAddItemTxUseCase(mockRepo, mockConfig)

		err := uc.Execute(ctx, roleID, 1001, 5, 0)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		if len(mockRepo.binaryData.BagData.Items) != 2 {
			t.Fatalf("Items count = %d, want 2", len(mockRepo.binaryData.BagData.Items))
		}
		// 第一个物品应该堆叠到10
		if mockRepo.binaryData.BagData.Items[0].Count != 10 {
			t.Errorf("First item Count = %d, want 10", mockRepo.binaryData.BagData.Items[0].Count)
		}
		// 第二个物品应该有4个
		if mockRepo.binaryData.BagData.Items[1].Count != 4 {
			t.Errorf("Second item Count = %d, want 4", mockRepo.binaryData.BagData.Items[1].Count)
		}
	})

	t.Run("背包已满时返回错误", func(t *testing.T) {
		items := make([]*protocol.ItemSt, 100)
		for i := range items {
			items[i] = &protocol.ItemSt{ItemId: uint32(1000 + i), Count: 1}
		}
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{
				BagData: &protocol.SiBagData{
					Items: items,
				},
			},
		}
		mockConfig := &mockConfigManager{
			itemConfig: map[uint32]interface{}{
				2001: &jsonconf.ItemConfig{
					ItemId:   2001,
					MaxStack: 1,
				},
			},
			bagConfig: map[uint32]interface{}{
				1: &jsonconf.BagConfig{Size: 100},
			},
		}
		uc := NewAddItemTxUseCase(mockRepo, mockConfig)

		err := uc.Execute(ctx, roleID, 2001, 1, 0)
		if err == nil {
			t.Fatal("Execute() should return error when bag is full")
		}
	})

	t.Run("物品配置不存在时返回错误", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{},
		}
		mockConfig := &mockConfigManager{
			itemConfig: map[uint32]interface{}{},
		}
		uc := NewAddItemTxUseCase(mockRepo, mockConfig)

		err := uc.Execute(ctx, roleID, 9999, 1, 0)
		if err == nil {
			t.Fatal("Execute() should return error when item config not found")
		}
	})

	t.Run("count为0时直接返回", func(t *testing.T) {
		mockRepo := &mockPlayerRepository{
			binaryData: &protocol.PlayerRoleBinaryData{},
		}
		mockConfig := &mockConfigManager{}
		uc := NewAddItemTxUseCase(mockRepo, mockConfig)

		err := uc.Execute(ctx, roleID, 1001, 0, 0)
		if err != nil {
			t.Fatalf("Execute() error = %v, want nil when count is 0", err)
		}
	})

	t.Run("Repository 返回错误时应该返回错误", func(t *testing.T) {
		testErr := errors.New("repository error")
		mockRepo := &mockPlayerRepository{
			err: testErr,
		}
		mockConfig := &mockConfigManager{
			itemConfig: map[uint32]interface{}{
				1001: &jsonconf.ItemConfig{ItemId: 1001, MaxStack: 1},
			},
		}
		uc := NewAddItemTxUseCase(mockRepo, mockConfig)

		err := uc.Execute(ctx, roleID, 1001, 1, 0)
		if err == nil {
			t.Fatal("Execute() should return error when repository fails")
		}
	})
}
