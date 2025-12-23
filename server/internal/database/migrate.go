package database

// AutoMigrate 所有表
func AutoMigrate() error {
	return DB.AutoMigrate(
		&Account{},
		&Player{},
		&ServerInfo{},
	)
}
