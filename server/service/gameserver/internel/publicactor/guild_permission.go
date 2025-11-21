package publicactor

import (
	"postapocgame/server/internal/protocol"
)

// 权限定义
const (
	// 说明：权限位的“第几位”由 proto/csproto/social_def.proto 中的 GuildPermission 枚举定义，
	// 这里仅通过 << 把枚举值映射为 bitmask，遵循文档第 7.3 节“共享枚举统一由 Proto 定义”的规范。

	// 副会长/会长可用的基础权限（按位）
	PermissionGuildApproveJoin = 1 << protocol.GuildPermission_GuildPermissionApproveJoin // 审批加入申请
	PermissionGuildKickMember  = 1 << protocol.GuildPermission_GuildPermissionKickMember  // 踢出成员
	PermissionGuildChangePos   = 1 << protocol.GuildPermission_GuildPermissionChangePos   // 修改职位（不能修改会长）
	PermissionGuildUpdateAnn   = 1 << protocol.GuildPermission_GuildPermissionUpdateAnn   // 修改宣言
	PermissionGuildUpdateName  = 1 << protocol.GuildPermission_GuildPermissionUpdateName  // 修改名称

	// 组长权限（仅限自己组）
	PermissionGuildApproveJoinGroup = 1 << protocol.GuildPermission_GuildPermissionApproveJoinGroup // 审批加入申请（仅限自己组）
	PermissionGuildKickMemberGroup  = 1 << protocol.GuildPermission_GuildPermissionKickMemberGroup  // 踢出成员（仅限自己组）

	// 成员权限
	PermissionGuildView = 1 << protocol.GuildPermission_GuildPermissionView // 查看公会信息

	// 会长权限：所有已定义的权限位
	PermissionGuildAll = PermissionGuildApproveJoin |
		PermissionGuildKickMember |
		PermissionGuildChangePos |
		PermissionGuildUpdateAnn |
		PermissionGuildUpdateName |
		PermissionGuildApproveJoinGroup |
		PermissionGuildKickMemberGroup |
		PermissionGuildView
)

// GetGuildPermission 获取公会职位对应的权限
func GetGuildPermission(position uint32) uint32 {
	switch protocol.GuildPosition(position) {
	case protocol.GuildPosition_GuildPositionLeader:
		return PermissionGuildAll
	case protocol.GuildPosition_GuildPositionViceLeader:
		return PermissionGuildApproveJoin | PermissionGuildKickMember | PermissionGuildChangePos | PermissionGuildUpdateAnn | PermissionGuildUpdateName | PermissionGuildView
	// case protocol.GuildPosition_GuildPositionGroupLeader:
	// 	// GroupLeader权限：审批加入申请（仅限自己组）、踢出成员（仅限自己组）、查看公会信息
	// 	return PermissionGuildApproveJoinGroup | PermissionGuildKickMemberGroup | PermissionGuildView
	// 注意：启用 GroupLeader 职位前，请确认 GuildPosition 与 GuildPermission 的 proto 枚举已生成并同步到客户端
	case protocol.GuildPosition_GuildPositionMember:
		return PermissionGuildView
	default:
		return 0
	}
}

// HasPermission 检查是否有指定权限
func HasPermission(position uint32, permission uint32) bool {
	userPermission := GetGuildPermission(position)
	return (userPermission & permission) != 0
}

// GetGuildMemberPosition 获取成员在公会中的职位
func GetGuildMemberPosition(guild *protocol.GuildData, roleId uint64) (uint32, bool) {
	for _, member := range guild.Members {
		if member.RoleId == roleId {
			return member.Position, true
		}
	}
	return 0, false
}

// CheckGuildPermission 检查操作者是否有权限执行操作
func CheckGuildPermission(guild *protocol.GuildData, operatorId uint64, requiredPermission uint32) bool {
	// 检查是否是成员
	position, ok := GetGuildMemberPosition(guild, operatorId)
	if !ok {
		return false
	}

	// 检查权限
	return HasPermission(position, requiredPermission)
}

// CanChangePosition 检查是否可以修改目标成员的职位
func CanChangePosition(guild *protocol.GuildData, operatorId uint64, targetId uint64, newPosition uint32) bool {
	operatorPos, ok1 := GetGuildMemberPosition(guild, operatorId)
	if !ok1 {
		return false
	}

	targetPos, ok2 := GetGuildMemberPosition(guild, targetId)
	if !ok2 {
		return false
	}

	// 会长可以修改任何人的职位（但不能修改自己的职位）
	if protocol.GuildPosition(operatorPos) == protocol.GuildPosition_GuildPositionLeader {
		return operatorId != targetId
	}

	// 副会长可以修改成员和组长的职位，但不能修改会长和副会长
	if protocol.GuildPosition(operatorPos) == protocol.GuildPosition_GuildPositionViceLeader {
		if protocol.GuildPosition(targetPos) == protocol.GuildPosition_GuildPositionLeader ||
			protocol.GuildPosition(targetPos) == protocol.GuildPosition_GuildPositionViceLeader {
			return false
		}
		// 不能将成员提升为会长或副会长
		if protocol.GuildPosition(newPosition) == protocol.GuildPosition_GuildPositionLeader ||
			protocol.GuildPosition(newPosition) == protocol.GuildPosition_GuildPositionViceLeader {
			return false
		}
		return true
	}

	return false
}
