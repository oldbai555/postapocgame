package playerrole

import (
	"context"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/playeractor/domain/model"
	"strings"
)

// CreateRoleInput 创建角色入参
type CreateRoleInput struct {
	AccountID uint64
	RoleName  string
	Job       uint32
	Sex       uint32
}

// CreateRoleResult 用例返回
type CreateRoleResult struct {
	Success bool
	Message string
	Role    *model.Role
}

// CreateRoleUseCase 创建角色用例
type CreateRoleUseCase struct {
	roleRepo iface.RoleRepository
}

// NewCreateRoleUseCase 创建用例
func NewCreateRoleUseCase(repo iface.RoleRepository) *CreateRoleUseCase {
	return &CreateRoleUseCase{roleRepo: repo}
}

// Execute 执行用例
func (uc *CreateRoleUseCase) Execute(ctx context.Context, input CreateRoleInput) (*CreateRoleResult, error) {
	roleName := strings.TrimSpace(input.RoleName)
	if roleName == "" {
		return &CreateRoleResult{
			Success: false,
			Message: "角色名不能为空",
		}, nil
	}

	if exists, err := uc.roleRepo.CheckRoleNameExists(ctx, roleName); err != nil {
		return nil, err
	} else if exists {
		return &CreateRoleResult{
			Success: false,
			Message: "角色名已存在",
		}, nil
	}

	roles, err := uc.roleRepo.GetRolesByAccountID(ctx, input.AccountID)
	if err != nil {
		return nil, err
	}
	if len(roles) >= 3 {
		return &CreateRoleResult{
			Success: false,
			Message: "角色数量已达上限",
		}, nil
	}

	role, err := uc.roleRepo.CreateRole(ctx, input.AccountID, roleName, input.Job, input.Sex)
	if err != nil {
		return nil, err
	}

	return &CreateRoleResult{
		Success: true,
		Message: "创建成功",
		Role:    role,
	}, nil
}
