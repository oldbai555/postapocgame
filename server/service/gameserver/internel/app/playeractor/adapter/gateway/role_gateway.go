package gateway

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/model"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"

	"gorm.io/gorm"

	"postapocgame/server/internal/database"
)

// RoleGateway 角色数据访问实现
type RoleGateway struct{}

// NewRoleGateway 创建角色 Gateway
func NewRoleGateway() repository.RoleRepository {
	return &RoleGateway{}
}

// GetRolesByAccountID 查询账号角色列表
func (g *RoleGateway) GetRolesByAccountID(_ context.Context, accountID uint64) ([]*model.Role, error) {
	dbRoles, err := database.GetPlayersByAccountID(uint(accountID))
	if err != nil {
		return nil, err
	}
	result := make([]*model.Role, 0, len(dbRoles))
	for _, r := range dbRoles {
		result = append(result, convertPlayer(r))
	}
	return result, nil
}

// CreateRole 创建角色
func (g *RoleGateway) CreateRole(_ context.Context, accountID uint64, roleName string, job, sex uint32) (*model.Role, error) {
	player, err := database.CreatePlayer(uint(accountID), roleName, int(job), int(sex))
	if err != nil {
		return nil, err
	}
	return convertPlayer(player), nil
}

// CheckRoleNameExists 检查角色名
func (g *RoleGateway) CheckRoleNameExists(_ context.Context, roleName string) (bool, error) {
	return database.CheckRoleNameExists(roleName)
}

// GetRoleByID 根据角色ID获取
func (g *RoleGateway) GetRoleByID(_ context.Context, roleID uint64) (*model.Role, error) {
	player, err := database.GetPlayerByID(uint(roleID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrRoleNotFound
		}
		return nil, err
	}
	return convertPlayer(player), nil
}

func convertPlayer(player *database.Player) *model.Role {
	if player == nil {
		return nil
	}
	return &model.Role{
		ID:        uint64(player.ID),
		AccountID: uint64(player.AccountID),
		RoleName:  player.RoleName,
		Job:       uint32(player.Job),
		Sex:       uint32(player.Sex),
		Level:     uint32(player.Level),
	}
}
