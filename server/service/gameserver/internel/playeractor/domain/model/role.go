package model

// Role 简化的角色信息
type Role struct {
	ID        uint64
	AccountID uint64
	RoleName  string
	Job       uint32
	Sex       uint32
	Level     uint32
}
