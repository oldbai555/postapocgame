package chat

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	chatdomain "postapocgame/server/service/gameserver/internel/domain/chat"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
	"time"
)

const (
	privateChatMaxRunes = 200
	privateChatCooldown = 10 * time.Second
)

// PrivateChatUseCase 私聊
type PrivateChatUseCase struct {
	configManager interfaces.ConfigManager
	publicGate    interfaces.PublicActorGateway
}

func NewPrivateChatUseCase(configManager interfaces.ConfigManager, publicGate interfaces.PublicActorGateway) *PrivateChatUseCase {
	return &PrivateChatUseCase{
		configManager: configManager,
		publicGate:    publicGate,
	}
}

func (uc *PrivateChatUseCase) Execute(
	ctx context.Context,
	limiter interfaces.ChatRateLimiter,
	roleID uint64,
	roleName string,
	targetID uint64,
	content string,
) error {
	if roleID == 0 {
		return customerr.NewError("未登录")
	}
	if targetID == 0 {
		return customerr.NewError("目标角色ID无效")
	}
	if roleID == targetID {
		return customerr.NewError("不能给自己发私聊")
	}
	if limiter == nil {
		return customerr.NewError("聊天系统未初始化")
	}
	if ok, reason := chatdomain.ValidateContent(content, privateChatMaxRunes); !ok {
		return customerr.NewError(reason)
	}
	if uc.containsSensitive(content) {
		return customerr.NewError("聊天内容包含敏感词，请重新输入")
	}

	now := servertime.Now()
	if !limiter.CanSend(now, privateChatCooldown) {
		return customerr.NewError("聊天过于频繁，请稍后再试")
	}
	limiter.MarkSent(now)

	msg := &protocol.ChatPrivateMsg{
		SenderId:   roleID,
		TargetId:   targetID,
		SenderName: roleName,
		Content:    content,
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return customerr.Wrap(err)
	}
	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdChatPrivate), data)
	return customerr.Wrap(uc.publicGate.SendMessageAsync(ctx, "global", actorMsg))
}

func (uc *PrivateChatUseCase) containsSensitive(content string) bool {
	words, ok := uc.configManager.GetSensitiveWordConfig()
	if !ok {
		return false
	}
	return chatdomain.ContainsSensitive(content, words)
}
