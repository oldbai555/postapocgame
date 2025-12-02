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

// QueryFriendListUseCase 查询好友列表
type QueryFriendListUseCase struct {
	playerRepo      repository.PlayerRepository
	publicActorGate interfaces.PublicActorGateway
}

// NewQueryFriendListUseCase 创建用例
func NewQueryFriendListUseCase(
	playerRepo repository.PlayerRepository,
	publicActor interfaces.PublicActorGateway,
) *QueryFriendListUseCase {
	return &QueryFriendListUseCase{
		playerRepo:      playerRepo,
		publicActorGate: publicActor,
	}
}

// Execute 触发好友列表查询
func (uc *QueryFriendListUseCase) Execute(ctx context.Context, roleID uint64, sessionID string) error {
	if roleID == 0 {
		return customerr.NewError("未登录")
	}

	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return customerr.Wrap(err)
	}
	friendData := frienddomain.EnsureFriendData(binaryData)
	if friendData == nil {
		return customerr.NewError("好友数据异常")
	}

	msg := &protocol.FriendListQueryMsg{
		RequesterId:        roleID,
		RequesterSessionId: sessionID,
		FriendIds:          append([]uint64(nil), friendData.FriendList...),
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return customerr.Wrap(err)
	}
	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdFriendListQuery), data)
	return uc.publicActorGate.SendMessageAsync(ctx, "global", actorMsg)
}
