package entitysystem

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/iface"
)

// AuctionSys 拍卖行系统
type AuctionSys struct {
	*BaseSystem
	data *protocol.SiAuctionData
}

// NewAuctionSys 创建拍卖行系统
func NewAuctionSys() iface.ISystem {
	return &AuctionSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysAuction)),
	}
}

func (s *AuctionSys) OnInit(ctx context.Context) {
	role, err := GetIPlayerRoleByContext(ctx)
	if err != nil || role == nil {
		return
	}
	bd := role.GetBinaryData()
	if bd.AuctionData == nil {
		bd.AuctionData = &protocol.SiAuctionData{
			AuctionIdList: make([]uint64, 0),
		}
	}
	s.data = bd.AuctionData
}

// GetAuctionIdList 获取拍卖ID列表
func (s *AuctionSys) GetAuctionIdList() []uint64 {
	if s.data == nil {
		return nil
	}
	return s.data.AuctionIdList
}

// AddAuctionId 添加拍卖ID
func (s *AuctionSys) AddAuctionId(auctionId uint64) bool {
	if s.data == nil {
		return false
	}
	// 检查是否已经存在
	for _, id := range s.data.AuctionIdList {
		if id == auctionId {
			return false
		}
	}
	s.data.AuctionIdList = append(s.data.AuctionIdList, auctionId)
	return true
}

// RemoveAuctionId 移除拍卖ID
func (s *AuctionSys) RemoveAuctionId(auctionId uint64) bool {
	if s.data == nil {
		return false
	}
	for i, id := range s.data.AuctionIdList {
		if id == auctionId {
			s.data.AuctionIdList = append(s.data.AuctionIdList[:i], s.data.AuctionIdList[i+1:]...)
			return true
		}
	}
	return false
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysAuction), NewAuctionSys)
}
