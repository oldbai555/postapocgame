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

// RespondFriendRequestUseCase 处理好友申请响应
type RespondFriendRequestUseCase struct {
	playerRepo      repository.PlayerRepository
	publicActorGate interfaces.PublicActorGateway
}

// NewRespondFriendRequestUseCase 创建用例
func NewRespondFriendRequestUseCase(
	playerRepo repository.PlayerRepository,
	publicActor interfaces.PublicActorGateway,
) *RespondFriendRequestUseCase {
	return &RespondFriendRequestUseCase{
		playerRepo:      playerRepo,
		publicActorGate: publicActor,
	}
}

// Execute 执行响应逻辑
func (uc *RespondFriendRequestUseCase) Execute(ctx context.Context, roleID uint64, requesterID uint64, accepted bool) (bool, string, error) {
	if roleID == 0 || requesterID == 0 {
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

	// 移除申请（若不存在则继续流程，保持兼容）
	friendData.FriendRequestList, _ = frienddomain.RemoveRequest(friendData.FriendRequestList, requesterID)

	if accepted {
		if !frienddomain.ContainsFriend(friendData.FriendList, requesterID) {
			friendData.FriendList = append(friendData.FriendList, requesterID)
		}

		resp := &protocol.AddFriendRespMsg{
			RequesterId: requesterID,
			TargetId:    roleID,
			Accepted:    true,
		}
		data, err := proto.Marshal(resp)
		if err != nil {
			return false, "", customerr.Wrap(err)
		}
		msg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAddFriendResp), data)
		if err := uc.publicActorGate.SendMessageAsync(ctx, "global", msg); err != nil {
			return false, "", customerr.Wrap(err)
		}
		return true, "已同意好友申请", nil
	}

	// 拒绝：也需要通知申请者
	resp := &protocol.AddFriendRespMsg{
		RequesterId: requesterID,
		TargetId:    roleID,
		Accepted:    false,
	}
	data, err := proto.Marshal(resp)
	if err != nil {
		return false, "", customerr.Wrap(err)
	}
	msg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAddFriendResp), data)
	if err := uc.publicActorGate.SendMessageAsync(ctx, "global", msg); err != nil {
		return false, "", customerr.Wrap(err)
	}
	return false, "已拒绝好友申请", nil
}
