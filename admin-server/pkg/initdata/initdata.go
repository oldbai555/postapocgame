package initdata

// IsInitUserID 检查是否是初始化用户ID
func IsInitUserID(id uint64) bool {
	return id == 1 // 超级管理员
}

// IsInitRoleID 检查是否是初始化角色ID
func IsInitRoleID(id uint64) bool {
	return id == 1 // 超级管理员角色
}

// IsInitPermissionID 检查是否是初始化权限ID
func IsInitPermissionID(id uint64) bool {
	// 初始化权限ID：1, 10-13, 20-23, 30-33, 40-43, 50-53, 60-63
	if id == 1 {
		return true
	}
	if id >= 10 && id <= 13 {
		return true
	}
	if id >= 20 && id <= 23 {
		return true
	}
	if id >= 30 && id <= 33 {
		return true
	}
	if id >= 40 && id <= 43 {
		return true
	}
	if id >= 50 && id <= 53 {
		return true
	}
	if id >= 60 && id <= 63 {
		return true
	}
	return false
}

// IsInitDepartmentID 检查是否是初始化部门ID
func IsInitDepartmentID(id uint64) bool {
	return id == 1 // 根部门
}

// IsInitMenuID 检查是否是初始化菜单ID
func IsInitMenuID(id uint64) bool {
	// 初始化菜单ID：1, 10-16
	return id == 1 || (id >= 10 && id <= 16)
}

// IsInitUserRoleID 检查是否是初始化用户-角色关联ID
func IsInitUserRoleID(id uint64) bool {
	return id == 1 // 超级管理员用户-角色关联
}

// IsInitRolePermissionID 检查是否是初始化角色-权限关联ID
func IsInitRolePermissionID(id uint64) bool {
	return id == 1 // 超级管理员角色-权限关联
}
