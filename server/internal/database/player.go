package database

import (
	"time"

	"postapocgame/server/internal/protocol"

	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

// Player 角色表
type Player struct {
	ID           uint   `gorm:"primaryKey"`
	AccountID    uint   `gorm:"not null;index"`
	RoleName     string `gorm:"not null;size:32;index"`
	Job          int
	Sex          int
	Level        int
	LastLoginAt  time.Time
	LastLogoutAt time.Time
	BinaryData   []byte `gorm:"type:blob"` // PlayerRoleBinaryData的二进制数据
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CreatePlayer 创建角色
func CreatePlayer(accountId uint, name string, job, sex int) (*Player, error) {
	player := &Player{AccountID: accountId, RoleName: name, Job: job, Sex: sex, Level: 1}
	result := DB.Create(player)
	return player, result.Error
}

// CheckRoleNameExists 检查角色名是否已存在
func CheckRoleNameExists(roleName string) (bool, error) {
	var count int64
	result := DB.Model(&Player{}).Where("role_name = ?", roleName).Count(&count)
	return count > 0, result.Error
}

// GetPlayersByAccountID 查询账号下所有角色
func GetPlayersByAccountID(accountId uint) ([]*Player, error) {
	var players []*Player
	result := DB.Where("account_id = ?", accountId).Find(&players)
	return players, result.Error
}

// GetPlayerByID 用角色ID查找
func GetPlayerByID(playerId uint) (*Player, error) {
	var player Player
	result := DB.First(&player, playerId)
	if result.Error != nil {
		return nil, result.Error
	}
	return &player, nil
}

// GetPlayerBinaryData 获取玩家的二进制数据
func GetPlayerBinaryData(playerId uint) (*protocol.PlayerRoleBinaryData, error) {
	player, err := GetPlayerByID(playerId)
	if err != nil {
		return nil, err
	}
	if len(player.BinaryData) == 0 {
		// 如果二进制数据为空，返回空的BinaryData
		return &protocol.PlayerRoleBinaryData{
			SysOpenStatus: make(map[uint32]uint32),
		}, nil
	}
	binaryData := &protocol.PlayerRoleBinaryData{}
	if err := proto.Unmarshal(player.BinaryData, binaryData); err != nil {
		return nil, err
	}
	if binaryData.SysOpenStatus == nil {
		binaryData.SysOpenStatus = make(map[uint32]uint32)
	}
	return binaryData, nil
}

// SavePlayerBinaryData 保存玩家的二进制数据
func SavePlayerBinaryData(playerId uint, binaryData *protocol.PlayerRoleBinaryData) error {
	if binaryData == nil {
		binaryData = &protocol.PlayerRoleBinaryData{
			SysOpenStatus: make(map[uint32]uint32),
		}
	}
	data, err := proto.Marshal(binaryData)
	if err != nil {
		return err
	}
	return DB.Model(&Player{}).Where("id = ?", playerId).Update("binary_data", data).Error
}

// SavePlayerBinaryDataTx 保存玩家的二进制数据（支持事务）
func SavePlayerBinaryDataTx(tx *gorm.DB, playerId uint, binaryData *protocol.PlayerRoleBinaryData) error {
	if binaryData == nil {
		binaryData = &protocol.PlayerRoleBinaryData{
			SysOpenStatus: make(map[uint32]uint32),
		}
	}
	data, err := proto.Marshal(binaryData)
	if err != nil {
		return err
	}
	return tx.Model(&Player{}).Where("id = ?", playerId).Update("binary_data", data).Error
}

// GetPlayerMainData 加载PlayerRoleMainData
func GetPlayerMainData(playerId uint) (*protocol.PlayerRoleMainData, error) {
	player, err := GetPlayerByID(playerId)
	if err != nil {
		return nil, err
	}
	mainData := &protocol.PlayerRoleMainData{
		RoleId:         uint64(player.ID),
		Job:            uint32(player.Job),
		Sex:            uint32(player.Sex),
		Level:          uint32(player.Level),
		RoleName:       player.RoleName,
		LastLoginTime:  player.LastLoginAt.Unix(),
		LastLogoutTime: player.LastLogoutAt.Unix(),
	}
	return mainData, nil
}

func UpdatePlayerLoginTime(playerId uint, loginAt time.Time) error {
	return DB.Model(&Player{}).Where("id = ?", playerId).Update("last_login_at", loginAt).Error
}

func UpdatePlayerLogoutTime(playerId uint, logoutAt time.Time) error {
	return DB.Model(&Player{}).Where("id = ?", playerId).Update("last_logout_at", logoutAt).Error
}
