package database

// TransactionAudit 交易审计表
type TransactionAudit struct {
	ID              uint   `gorm:"primaryKey"`
	TransactionType uint32 `gorm:"not null;index"`     // 交易类型：1=拍卖行购买, 2=拍卖行出售, 3=其他
	BuyerId         uint64 `gorm:"not null;index"`     // 买家ID
	SellerId        uint64 `gorm:"not null;index"`     // 卖家ID
	ItemId          uint32 `gorm:"not null"`           // 物品ID
	Count           uint32 `gorm:"not null"`           // 数量
	Price           int64  `gorm:"not null"`           // 价格
	Status          uint32 `gorm:"not null;default:1"` // 状态：1=成功, 2=失败, 3=可疑
	Reason          string `gorm:"type:text"`          // 备注/原因
	CreatedAt       int64  `gorm:"autoCreateTime"`
	UpdatedAt       int64  `gorm:"autoUpdateTime"`
}

// SaveTransactionAudit 保存交易审计记录
func SaveTransactionAudit(transactionType uint32, buyerId uint64, sellerId uint64, itemId uint32, count uint32, price int64, status uint32, reason string) error {
	audit := &TransactionAudit{
		TransactionType: transactionType,
		BuyerId:         buyerId,
		SellerId:        sellerId,
		ItemId:          itemId,
		Count:           count,
		Price:           price,
		Status:          status,
		Reason:          reason,
	}
	return DB.Create(audit).Error
}

// GetTransactionAuditByRoleId 获取角色的交易审计记录
func GetTransactionAuditByRoleId(roleId uint64, limit int) ([]*TransactionAudit, error) {
	var audits []TransactionAudit
	query := DB.Where("buyer_id = ? OR seller_id = ?", roleId, roleId).
		Order("created_at DESC").
		Limit(limit)
	result := query.Find(&audits)
	if result.Error != nil {
		return nil, result.Error
	}

	resultList := make([]*TransactionAudit, 0, len(audits))
	for i := range audits {
		resultList = append(resultList, &audits[i])
	}
	return resultList, nil
}

// GetSuspiciousTransactions 获取可疑交易记录
func GetSuspiciousTransactions(limit int) ([]*TransactionAudit, error) {
	var audits []TransactionAudit
	query := DB.Where("status = ?", 3).
		Order("created_at DESC").
		Limit(limit)
	result := query.Find(&audits)
	if result.Error != nil {
		return nil, result.Error
	}

	resultList := make([]*TransactionAudit, 0, len(audits))
	for i := range audits {
		resultList = append(resultList, &audits[i])
	}
	return resultList, nil
}
