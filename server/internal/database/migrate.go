package database

// AutoMigrate 所有表
func AutoMigrate() error {
	return DB.AutoMigrate(
		&Account{},
		&Player{},
		&OfflineMessage{},
		&OfflineData{},
		&PlayerActorMessage{},
		// &Guild{},        // 已移除：公会系统已删除
		// &AuctionItem{},  // 已移除：拍卖行系统已删除
		&Blacklist{}, // 保留：聊天系统仍在使用黑名单功能
		// &TransactionAudit{}, // 已移除：拍卖行交易审计功能已删除
		&ServerInfo{},
	)
}
