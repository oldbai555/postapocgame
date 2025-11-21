package entitysystem

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/iface"
)

// FriendSys 好友系统
type FriendSys struct {
	*BaseSystem
	data *protocol.SiFriendData
}

// NewFriendSys 创建好友系统
func NewFriendSys() iface.ISystem {
	return &FriendSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysFriend)),
	}
}

func (s *FriendSys) OnInit(ctx context.Context) {
	role, err := GetIPlayerRoleByContext(ctx)
	if err != nil || role == nil {
		return
	}
	bd := role.GetBinaryData()
	if bd.FriendData == nil {
		bd.FriendData = &protocol.SiFriendData{
			FriendList:        make([]uint64, 0),
			FriendRequestList: make([]uint64, 0),
		}
	}
	s.data = bd.FriendData
}

// GetFriendList 获取好友列表
func (s *FriendSys) GetFriendList() []uint64 {
	if s.data == nil {
		return nil
	}
	return s.data.FriendList
}

// GetFriendRequestList 获取好友申请列表
func (s *FriendSys) GetFriendRequestList() []uint64 {
	if s.data == nil {
		return nil
	}
	return s.data.FriendRequestList
}

// AddFriend 添加好友
func (s *FriendSys) AddFriend(friendId uint64) bool {
	if s.data == nil {
		return false
	}
	// 检查是否已经是好友
	for _, id := range s.data.FriendList {
		if id == friendId {
			return false
		}
	}
	s.data.FriendList = append(s.data.FriendList, friendId)
	return true
}

// RemoveFriend 移除好友
func (s *FriendSys) RemoveFriend(friendId uint64) bool {
	if s.data == nil {
		return false
	}
	for i, id := range s.data.FriendList {
		if id == friendId {
			s.data.FriendList = append(s.data.FriendList[:i], s.data.FriendList[i+1:]...)
			return true
		}
	}
	return false
}

// AddFriendRequest 添加好友申请
func (s *FriendSys) AddFriendRequest(requesterId uint64) bool {
	if s.data == nil {
		return false
	}
	// 检查是否已经申请过
	for _, id := range s.data.FriendRequestList {
		if id == requesterId {
			return false
		}
	}
	s.data.FriendRequestList = append(s.data.FriendRequestList, requesterId)
	return true
}

// RemoveFriendRequest 移除好友申请
func (s *FriendSys) RemoveFriendRequest(requesterId uint64) bool {
	if s.data == nil {
		return false
	}
	for i, id := range s.data.FriendRequestList {
		if id == requesterId {
			s.data.FriendRequestList = append(s.data.FriendRequestList[:i], s.data.FriendRequestList[i+1:]...)
			return true
		}
	}
	return false
}

// GetFriendSys 获取好友系统
func GetFriendSys(ctx context.Context) *FriendSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysFriend))
	if system == nil {
		return nil
	}
	sys := system.(*FriendSys)
	if sys == nil || !sys.IsOpened() {
		return nil
	}
	return sys
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysFriend), NewFriendSys)
}
