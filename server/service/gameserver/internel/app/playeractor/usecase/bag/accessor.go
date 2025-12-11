package bag

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
)

// accessor 负责集中处理背包数据的索引/增删查，避免各处重复构建 map。
type accessor struct {
	bagData *protocol.SiBagData
	index   map[uint32][]*protocol.ItemSt
}

// Snapshot 是对 accessor 的对外只读封装，便于 SystemAdapter 查询或做快照。
type Snapshot struct {
	*accessor
}

// NewBagSnapshot 构造只读视图；修改操作仍作用于同一份 BinaryData。
func NewBagSnapshot(ctx context.Context, repo repository.PlayerRepository) (*Snapshot, error) {
	acc, err := newAccessor(ctx, repo)
	if err != nil {
		return nil, err
	}
	return &Snapshot{accessor: acc}, nil
}

func (s *Snapshot) Find(itemID uint32, bind uint32) *protocol.ItemSt {
	if bind == 0 {
		if items, ok := s.index[itemID]; ok && len(items) > 0 {
			return items[0]
		}
		return nil
	}
	return s.findStackable(itemID, bind)
}

func (s *Snapshot) Total(itemID uint32) uint32 {
	return s.totalCount(itemID)
}

func (s *Snapshot) Snapshot() map[uint32]*protocol.ItemSt {
	return s.snapshot()
}

func (s *Snapshot) Restore(snapshot map[uint32]*protocol.ItemSt) {
	s.restore(snapshot)
}

func newAccessor(ctx context.Context, repo repository.PlayerRepository) (*accessor, error) {
	bagData, err := repo.GetBagData(ctx)
	if err != nil {
		return nil, err
	}
	acc := &accessor{
		bagData: bagData,
		index:   make(map[uint32][]*protocol.ItemSt),
	}
	acc.rebuildIndex()
	return acc, nil
}

func (a *accessor) rebuildIndex() {
	a.index = make(map[uint32][]*protocol.ItemSt)
	if a.bagData == nil {
		return
	}
	for _, item := range a.bagData.Items {
		if item != nil {
			a.index[item.ItemId] = append(a.index[item.ItemId], item)
		}
	}
}

func (a *accessor) totalCount(itemID uint32) uint32 {
	var total uint32
	if items, ok := a.index[itemID]; ok {
		for _, item := range items {
			if item != nil {
				total += item.Count
			}
		}
	}
	return total
}

func (a *accessor) findStackable(itemID uint32, bind uint32) *protocol.ItemSt {
	if items, ok := a.index[itemID]; ok {
		for _, item := range items {
			if item != nil && item.Bind == bind {
				return item
			}
		}
	}
	return nil
}

func (a *accessor) addItem(itemID, bind, count, maxStack, bagSize uint32) error {
	if count == 0 {
		return nil
	}
	if maxStack > 1 {
		if existing := a.findStackable(itemID, bind); existing != nil {
			maxAdd := maxStack - existing.Count
			if maxAdd > 0 {
				addCount := count
				if addCount > maxAdd {
					addCount = maxAdd
				}
				existing.Count += addCount
				count -= addCount
			}
		}
	}

	if count > 0 {
		if len(a.bagData.Items) >= int(bagSize) {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag is full")
		}
		newItem := &protocol.ItemSt{
			ItemId: itemID,
			Count:  count,
			Bind:   bind,
		}
		a.bagData.Items = append(a.bagData.Items, newItem)
	}

	a.rebuildIndex()
	return nil
}

func (a *accessor) removeItem(itemID, count uint32) error {
	if count == 0 {
		return nil
	}
	remaining := count
	for i := 0; i < len(a.bagData.Items) && remaining > 0; {
		item := a.bagData.Items[i]
		if item == nil || item.ItemId != itemID {
			i++
			continue
		}
		if item.Count > remaining {
			item.Count -= remaining
			remaining = 0
			break
		}
		remaining -= item.Count
		a.bagData.Items = append(a.bagData.Items[:i], a.bagData.Items[i+1:]...)
		continue
	}
	if remaining > 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not enough")
	}
	a.rebuildIndex()
	return nil
}

func (a *accessor) snapshot() map[uint32]*protocol.ItemSt {
	snapshot := make(map[uint32]*protocol.ItemSt)
	for _, item := range a.bagData.Items {
		if item == nil {
			continue
		}
		key := item.ItemId*1000 + item.Bind
		snapshot[key] = &protocol.ItemSt{
			ItemId: item.ItemId,
			Count:  item.Count,
			Bind:   item.Bind,
		}
	}
	return snapshot
}

func (a *accessor) restore(snapshot map[uint32]*protocol.ItemSt) {
	a.bagData.Items = a.bagData.Items[:0]
	for _, item := range snapshot {
		if item == nil {
			continue
		}
		a.bagData.Items = append(a.bagData.Items, &protocol.ItemSt{
			ItemId: item.ItemId,
			Count:  item.Count,
			Bind:   item.Bind,
		})
	}
	a.rebuildIndex()
}
