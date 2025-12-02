package friend

import "postapocgame/server/internal/protocol"

// EnsureFriendData 确保 FriendData 已初始化
func EnsureFriendData(binaryData *protocol.PlayerRoleBinaryData) *protocol.SiFriendData {
	if binaryData == nil {
		return nil
	}
	if binaryData.FriendData == nil {
		binaryData.FriendData = &protocol.SiFriendData{
			FriendList:        make([]uint64, 0),
			FriendRequestList: make([]uint64, 0),
		}
	}
	return binaryData.FriendData
}

// ContainsFriend 判断是否已有好友
func ContainsFriend(list []uint64, target uint64) bool {
	for _, id := range list {
		if id == target {
			return true
		}
	}
	return false
}

// RemoveFriend 从列表移除好友
func RemoveFriend(list []uint64, target uint64) ([]uint64, bool) {
	for i, id := range list {
		if id == target {
			return append(list[:i], list[i+1:]...), true
		}
	}
	return list, false
}

// RemoveRequest 从申请列表移除
func RemoveRequest(list []uint64, target uint64) ([]uint64, bool) {
	for i, id := range list {
		if id == target {
			return append(list[:i], list[i+1:]...), true
		}
	}
	return list, false
}
