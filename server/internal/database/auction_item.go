package database

import (
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"postapocgame/server/internal/protocol"
)

// AuctionItem 拍卖物品表
type AuctionItem struct {
	ID         uint   `gorm:"primaryKey"`
	AuctionId  uint64 `gorm:"not null;uniqueIndex"`
	ItemId     uint32 `gorm:"not null;index"`
	Count      uint32 `gorm:"not null"`
	Price      int64  `gorm:"not null"`
	SellerId   uint64 `gorm:"not null;index"`
	ExpireTime int64  `gorm:"not null;index"`
	CreateTime int64  `gorm:"not null"`
	BinaryData []byte `gorm:"type:blob"` // AuctionItem的二进制数据
	CreatedAt  int64  `gorm:"autoCreateTime"`
	UpdatedAt  int64  `gorm:"autoUpdateTime"`
}

// SaveAuctionItem 保存拍卖物品
func SaveAuctionItem(auctionItem *protocol.AuctionItem) error {
	data, err := proto.Marshal(auctionItem)
	if err != nil {
		return err
	}

	item := &AuctionItem{
		AuctionId:  auctionItem.AuctionId,
		ItemId:     auctionItem.ItemId,
		Count:      auctionItem.Count,
		Price:      auctionItem.Price,
		SellerId:   auctionItem.SellerId,
		ExpireTime: auctionItem.ExpireTime,
		CreateTime: auctionItem.CreateTime,
		BinaryData: data,
	}

	// 使用AuctionId作为唯一键，如果存在则更新，否则创建
	var existingItem AuctionItem
	result := DB.Where("auction_id = ?", auctionItem.AuctionId).First(&existingItem)
	if result.Error == gorm.ErrRecordNotFound {
		// 创建新记录
		return DB.Create(item).Error
	} else if result.Error != nil {
		return result.Error
	} else {
		// 更新现有记录
		return DB.Model(&existingItem).Updates(item).Error
	}
}

// GetAuctionItem 获取拍卖物品
func GetAuctionItem(auctionId uint64) (*protocol.AuctionItem, error) {
	var item AuctionItem
	result := DB.Where("auction_id = ?", auctionId).First(&item)
	if result.Error != nil {
		return nil, result.Error
	}

	auctionItem := &protocol.AuctionItem{}
	if err := proto.Unmarshal(item.BinaryData, auctionItem); err != nil {
		return nil, err
	}

	return auctionItem, nil
}

// GetAllAuctionItems 获取所有拍卖物品
func GetAllAuctionItems() ([]*protocol.AuctionItem, error) {
	var items []AuctionItem
	result := DB.Find(&items)
	if result.Error != nil {
		return nil, result.Error
	}

	auctionItems := make([]*protocol.AuctionItem, 0, len(items))
	for _, item := range items {
		auctionItem := &protocol.AuctionItem{}
		if err := proto.Unmarshal(item.BinaryData, auctionItem); err != nil {
			continue
		}
		auctionItems = append(auctionItems, auctionItem)
	}

	return auctionItems, nil
}

// DeleteAuctionItem 删除拍卖物品
func DeleteAuctionItem(auctionId uint64) error {
	return DB.Where("auction_id = ?", auctionId).Delete(&AuctionItem{}).Error
}

// CleanExpiredAuctionItems 清理过期的拍卖物品
func CleanExpiredAuctionItems(now int64) error {
	return DB.Where("expire_time < ?", now).Delete(&AuctionItem{}).Error
}
