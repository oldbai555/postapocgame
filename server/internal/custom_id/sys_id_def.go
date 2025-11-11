/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package custom_id

// 系统Id
type SystemId uint32

const (
	SysQuest      SystemId = 1 // 任务系统
	SysLevel      SystemId = 2 // 等级系统
	SysBag        SystemId = 3 // 背包系统
	SysVip        SystemId = 4 // VIP系统
	SysMoney      SystemId = 5 // 货币系统
	SysAttr       SystemId = 6 // 属性系统
	SysOfflineMsg SystemId = 7 // 离线消息系统
	SysMail       SystemId = 8 // 邮件系统

	SysIdMax SystemId = 9 // 最大系统Id
)
