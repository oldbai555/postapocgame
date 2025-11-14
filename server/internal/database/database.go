package database

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Init 初始化数据库（仅支持sqlite）
func Init(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}
	return nil
}
