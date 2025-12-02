package vip

import "postapocgame/server/internal/protocol"

// EnsureVipData 确保 VipData 初始化
func EnsureVipData(binaryData *protocol.PlayerRoleBinaryData) *protocol.SiVipData {
	if binaryData == nil {
		return nil
	}
	if binaryData.VipData == nil {
		binaryData.VipData = &protocol.SiVipData{
			Level: 0,
			Exp:   0,
		}
	}
	return binaryData.VipData
}
