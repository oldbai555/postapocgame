package friend

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	frienddomain "postapocgame/server/service/gameserver/internel/domain/friend"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// SendFriendRequestUseCase 发送好友申请
type SendFriendRequestUseCase struct {
	playerRepo      repository.PlayerRepository
	publicActorGate interfaces.PublicActorGateway
	blacklistRepo   interfaces.BlacklistRepository
}

// NewSendFriendRequestUseCase 创建用例
func NewSendFriendRequestUseCase(
	playerRepo repository.PlayerRepository,
	publicActor interfaces.PublicActorGateway,
	blacklistRepo interfaces.BlacklistRepository,
) *SendFriendRequestUseCase {
	return &SendFriendRequestUseCase{
		playerRepo:      playerRepo,
		publicActorGate: publicActor,
		blacklistRepo:   blacklistRepo,
	}
}

// Execute 执行发送逻辑
func (uc *SendFriendRequestUseCase) Execute(ctx context.Context, roleID uint64, roleName string, targetID uint64) error {
	if roleID == 0 {
		return customerr.NewError("invalid role id")
	}
	if targetID == 0 {
		return customerr.NewError("invalid target id")
	}
	if roleID == targetID {
		return customerr.NewError("不能添加自己为好友")
	}

	// 检查黑名单：若自己拉黑了目标或被目标拉黑，则拒绝
	if uc.blacklistRepo != nil {
		if blocked, err := uc.blacklistRepo.IsBlocked(ctx, roleID, targetID); err == nil && blocked {
			return customerr.NewError("已将该玩家加入黑名单")
		}
		if blocked, err := uc.blacklistRepo.IsBlocked(ctx, targetID, roleID); err == nil && blocked {
			return customerr.NewError("对方已将你加入黑名单")
		}
	}

	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return customerr.Wrap(err)
	}
	friendData := frienddomain.EnsureFriendData(binaryData)
	if friendData == nil {
		return customerr.NewError("好友数据异常")
	}
	if frienddomain.ContainsFriend(friendData.FriendList, targetID) {
		return customerr.NewError("已经是好友")
	}

	req := &protocol.AddFriendReqMsg{
		RequesterId:   roleID,
		TargetId:      targetID,
		RequesterName: roleName,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		return customerr.Wrap(err)
	}

	msg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAddFriendReq), data)
	if err := uc.publicActorGate.SendMessageAsync(ctx, "global", msg); err != nil {
		return customerr.Wrap(err)
	}
	return nil
}
