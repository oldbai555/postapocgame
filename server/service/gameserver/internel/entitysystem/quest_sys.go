package entitysystem

import (
	"fmt"
	"postapocgame/server/internal/custom_id"
	protocol2 "postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/iface"
)

var (
	ErrQuestNotFound        = fmt.Errorf("quest not found")
	ErrQuestAlreadyAccepted = fmt.Errorf("quest already accepted")
	ErrQuestNotComplete     = fmt.Errorf("quest not complete")
)

// QuestSys 任务系统
type QuestSys struct {
	*BaseSystem
	quests map[uint32]*protocol2.Quest // questID -> Quest
}

// NewQuestSys 创建任务系统
func NewQuestSys(role iface.IPlayerRole) *QuestSys {
	sys := &QuestSys{
		BaseSystem: NewBaseSystem(custom_id.SysQuest, role),
		quests:     make(map[uint32]*protocol2.Quest),
	}
	return sys
}

// OnRoleLogin 角色登录时下发任务数据
func (s *QuestSys) OnRoleLogin() {
	return
}

// SendData 下发任务数据
func (s *QuestSys) SendData() error {
	quests := make([]protocol2.Quest, 0, len(s.quests))
	for _, quest := range s.quests {
		quests = append(quests, *quest)
	}

	data := &protocol2.QuestData{
		Quests: quests,
	}
	jsonData, _ := tool.JsonMarshal(data)
	return s.role.SendMessage(1, 5, jsonData)
}

// AcceptQuest 接取任务
func (s *QuestSys) AcceptQuest(questId uint32) error {
	if _, ok := s.quests[questId]; ok {
		return ErrQuestAlreadyAccepted
	}

	quest := &protocol2.Quest{
		QuestId:  questId,
		Progress: 0,
		Status:   1, // 进行中
	}

	s.quests[questId] = quest
	return s.SendData()
}

// UpdateProgress 更新任务进度
func (s *QuestSys) UpdateProgress(questId uint32, progress uint32) error {
	quest, ok := s.quests[questId]
	if !ok {
		return ErrQuestNotFound
	}

	quest.Progress = progress

	// TODO: 检查是否完成
	// if quest.Progress >= questConfig.Target {
	//     quest.Status = 2 // 已完成
	// }

	return s.SendData()
}

// CompleteQuest 完成任务（领取奖励）
func (s *QuestSys) CompleteQuest(questID uint32) error {
	quest, ok := s.quests[questID]
	if !ok {
		return ErrQuestNotFound
	}

	if quest.Status != 2 {
		return ErrQuestNotComplete
	}

	quest.Status = 3 // 已领取

	// TODO: 发放奖励
	// s.playerrole.GiveAwards(questConfig.Rewards)

	return s.SendData()
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(custom_id.SysQuest, func(role iface.IPlayerRole) iface.ISystem {
		return NewQuestSys(role)
	})
}
