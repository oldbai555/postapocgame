/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package event

const (
	EventRoleLogin  Type = iota + 1 // 角色登录
	EventRoleLogout                 // 角色登出
	EventRoleUpLv                   // 角色升级
	EventAddExp                     // 增加经验
	EventAddMoney                   // 增加货币
	EventAddItem                    // 增加道具
)
