package gevent

import "postapocgame/server/internal/event"

// 服务器级别事件
const (
	OnSrvStart event.Type = iota + 1
)

// 玩家级别事件（从1000开始，避免冲突）
const (
	// 玩家登录相关
	OnPlayerLogin event.Type = iota + 1000
	OnPlayerLogout
	OnPlayerReconnect

	// 等级系统
	OnPlayerLevelUp
	OnPlayerExpChange

	// 货币系统
	OnMoneyChange
	OnGoldChange
	OnDiamondChange
	OnCoinChange

	// 背包系统
	OnItemAdd
	OnItemRemove
	OnItemUse
	OnBagExpand

	// 装备系统
	OnEquipChange
	OnEquipUpgrade

	// 任务系统
	OnQuestAccept
	OnQuestProgress
	OnQuestComplete

	// 邮件系统
	OnMailReceive
	OnMailRead
	OnMailDelete
	OnMailRewardReceived

	// VIP系统
	OnVipLevelUp
	OnVipExpChange
)
