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

func (pr *PlayerRole) GetLevelData() *protocol.SiLevelData {
	data := pr.GetBinaryData()
	if data.LevelData == nil {
		data.LevelData = &protocol.SiLevelData{}
	}
	return data.LevelData
}

func (pr *PlayerRole) GetSkillData() *protocol.SiSkillData {
	data := pr.GetBinaryData()
	if data.SkillData == nil {
		data.SkillData = &protocol.SiSkillData{}
	}
	return data.SkillData
}
