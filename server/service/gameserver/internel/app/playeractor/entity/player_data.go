/**
 * @Author: zjj
 * @Date: 2025/12/10
 * @Desc:
**/

package entity

import "postapocgame/server/internal/protocol"

func (pr *PlayerRole) GetBinaryData() *protocol.PlayerRoleBinaryData {
	return pr.BinaryData
}

func (pr *PlayerRole) GetBagData() *protocol.SiBagData {
	data := pr.GetBinaryData()
	if data.BagData == nil {
		data.BagData = &protocol.SiBagData{}
	}
	return data.BagData
}

func (pr *PlayerRole) GetMoneyData() *protocol.SiMoneyData {
	data := pr.GetBinaryData()
	if data.MoneyData == nil {
		data.MoneyData = &protocol.SiMoneyData{}
	}
	if data.MoneyData.MoneyMap == nil {
		data.MoneyData.MoneyMap = make(map[uint32]int64)
	}
	return data.MoneyData
}

func (pr *PlayerRole) GetLevelData() *protocol.SiLevelData {
	data := pr.GetBinaryData()
	if data.LevelData == nil {
		data.LevelData = &protocol.SiLevelData{}
	}
	return data.LevelData
}

func (pr *PlayerRole) GetEquipData() *protocol.SiEquipData {
	data := pr.GetBinaryData()
	if data.EquipData == nil {
		data.EquipData = &protocol.SiEquipData{}
	}
	return data.EquipData
}

func (pr *PlayerRole) GetSkillData() *protocol.SiSkillData {
	data := pr.GetBinaryData()
	if data.SkillData == nil {
		data.SkillData = &protocol.SiSkillData{}
	}
	return data.SkillData
}

func (pr *PlayerRole) GetItemUseData() *protocol.SiItemUseData {
	data := pr.GetBinaryData()
	if data.ItemUseData == nil {
		data.ItemUseData = &protocol.SiItemUseData{}
	}
	if data.ItemUseData.CooldownMap == nil {
		data.ItemUseData.CooldownMap = make(map[uint32]int64)
	}
	return data.ItemUseData
}

func (pr *PlayerRole) GetQuestData() *protocol.SiQuestData {
	data := pr.GetBinaryData()
	if data.QuestData == nil {
		data.QuestData = &protocol.SiQuestData{}
	}
	if data.QuestData.QuestMap == nil {
		data.QuestData.QuestMap = make(map[uint32]*protocol.QuestTypeList)
	}
	if data.QuestData.LastResetMap == nil {
		data.QuestData.LastResetMap = make(map[uint32]int64)
	}
	if data.QuestData.QuestFinishCount == nil {
		data.QuestData.QuestFinishCount = make(map[uint32]uint32)
	}
	return data.QuestData
}

func (pr *PlayerRole) GetDungeonData() *protocol.SiDungeonData {
	data := pr.GetBinaryData()
	if data.DungeonData == nil {
		data.DungeonData = &protocol.SiDungeonData{}
	}
	return data.DungeonData
}

func (pr *PlayerRole) GetMailData() *protocol.SiMailData {
	data := pr.GetBinaryData()
	if data.MailData == nil {
		data.MailData = &protocol.SiMailData{}
	}
	return data.MailData
}
