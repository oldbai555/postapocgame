package playerrole

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/model"
	"postapocgame/server/service/gameserver/internel/iface"
)

// QueryRolesResult 角色列表结果
type QueryRolesResult struct {
	Roles []*model.Role
}

// QueryRolesUseCase 查询账号角色列表
type QueryRolesUseCase struct {
	roleRepo iface.RoleRepository
}

// NewQueryRolesUseCase 创建用例
func NewQueryRolesUseCase(repo iface.RoleRepository) *QueryRolesUseCase {
	return &QueryRolesUseCase{
		roleRepo: repo,
	}
}

// Execute 执行查询
func (uc *QueryRolesUseCase) Execute(ctx context.Context, accountID uint64) (*QueryRolesResult, error) {
	roles, err := uc.roleRepo.GetRolesByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return &QueryRolesResult{
		Roles: roles,
	}, nil
}
