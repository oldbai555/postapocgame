package systems

import (
	"time"

	"postapocgame/server/example/internal/client"
	"postapocgame/server/internal/protocol"
)

const defaultWait = 8 * time.Second

type InventorySystem struct {
	core *client.Core
}

func NewInventorySystem(core *client.Core) *InventorySystem {
	return &InventorySystem{core: core}
}

func (s *InventorySystem) Refresh(timeout time.Duration) ([]*protocol.ItemSt, error) {
	if err := s.core.RequestBagData(); err != nil {
		return nil, err
	}
	if _, err := s.core.WaitBagData(timeout); err != nil {
		return nil, err
	}
	return s.core.BagSnapshot(), nil
}

func (s *InventorySystem) UseItem(itemID, count uint32, timeout time.Duration) (*protocol.S2CUseItemResultReq, error) {
	if err := s.core.UseItem(itemID, count); err != nil {
		return nil, err
	}
	return s.core.WaitUseItemResult(timeout)
}

func (s *InventorySystem) Pickup(handle uint64, timeout time.Duration) (*protocol.S2CPickupItemResultReq, error) {
	if err := s.core.PickupItem(handle); err != nil {
		return nil, err
	}
	return s.core.WaitPickupResult(timeout)
}
