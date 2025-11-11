/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package jsonconf

import "postapocgame/server/internal/protocol"

// QuestConfig 任务配置
type QuestConfig struct {
	QuestId     uint32           `json:"questId"`     // 任务Id
	Name        string           `json:"name"`        // 任务名称
	Type        uint32           `json:"type"`        // 任务类型: 1=主线 2=支线 3=日常
	Description string           `json:"description"` // 任务描述
	Level       uint32           `json:"level"`       // 需求等级
	PreQuests   []uint32         `json:"preQuests"`   // 前置任务
	Objectives  []QuestObjective `json:"objectives"`  // 任务目标
	Rewards     []protocol.Item  `json:"rewards"`     // 任务奖励
	ExpReward   uint64           `json:"expReward"`   // 经验奖励
}

// QuestObjective 任务目标
type QuestObjective struct {
	Type       uint32 `json:"type"`       // 目标类型: 1=击杀怪物 2=收集道具 3=与NPC对话
	TargetId   uint32 `json:"targetId"`   // 目标Id
	TargetName string `json:"targetName"` // 目标名称
	Count      uint32 `json:"count"`      // 目标数量
}
