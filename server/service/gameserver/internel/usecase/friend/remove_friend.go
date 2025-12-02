package friend

import (
	"context"
	"postapocgame/server/pkg/customerr"
	frienddomain "postapocgame/server/service/gameserver/internel/domain/friend"
	"postapocgame/server/service/gameserver/internel/domain/repository"
)

// RemoveFriendUseCase 移除好友
type RemoveFriendUseCase struct {
	playerRepo repository.PlayerRepository
}

// NewRemoveFriendUseCase 创建用例
func NewRemoveFriendUseCase(playerRepo repository.PlayerRepository) *RemoveFriendUseCase {
	return &RemoveFriendUseCase{playerRepo: playerRepo}
}

// Execute 移除好友
func (uc *RemoveFriendUseCase) Execute(ctx context.Context, roleID uint64, friendID uint64) (bool, string, error) {
	if roleID == 0 || friendID == 0 {
		return false, "", customerr.NewError("参数错误")
	}
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return false, "", customerr.Wrap(err)
	}
	friendData := frienddomain.EnsureFriendData(binaryData)
	if friendData == nil {
		return false, "", customerr.NewError("好友数据异常")
	}
	updated, removed := frienddomain.RemoveFriend(friendData.FriendList, friendID)
	friendData.FriendList = updated
	if !removed {
		return false, "未找到该好友", nil
	}
	return true, "删除成功", nil
}
