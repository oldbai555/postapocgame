package auction

import "postapocgame/server/internal/protocol"

// EnsureAuctionData 确保拍卖数据初始化
func EnsureAuctionData(binaryData *protocol.PlayerRoleBinaryData) *protocol.SiAuctionData {
	if binaryData == nil {
		return nil
	}
	if binaryData.AuctionData == nil {
		binaryData.AuctionData = &protocol.SiAuctionData{
			AuctionIdList: make([]uint64, 0),
		}
	}
	return binaryData.AuctionData
}
