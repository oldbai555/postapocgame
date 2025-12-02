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
	worldChatMaxRunes = 200
	worldChatCooldown = 10 * time.Second
)

// WorldChatUseCase 世界聊天
type WorldChatUseCase struct {
	configManager interfaces.ConfigManager
	publicGate    interfaces.PublicActorGateway
}

func NewWorldChatUseCase(configManager interfaces.ConfigManager, publicGate interfaces.PublicActorGateway) *WorldChatUseCase {
	return &WorldChatUseCase{
		configManager: configManager,
		publicGate:    publicGate,
	}
}

func (uc *WorldChatUseCase) Execute(
	ctx context.Context,
	limiter interfaces.ChatRateLimiter,
	roleID uint64,
	roleName string,
	content string,
) error {
	if roleID == 0 {
		return customerr.NewError("未登录")
	}
	if limiter == nil {
		return customerr.NewError("聊天系统未初始化")
	}
	if ok, reason := chatdomain.ValidateContent(content, worldChatMaxRunes); !ok {
		return customerr.NewError(reason)
	}
	if uc.containsSensitive(content) {
		return customerr.NewError("聊天内容包含敏感词，请重新输入")
	}
	now := servertime.Now()
	if !limiter.CanSend(now, worldChatCooldown) {
		return customerr.NewError("聊天过于频繁，请稍后再试")
	}
	limiter.MarkSent(now)

	msg := &protocol.ChatWorldMsg{
		SenderId:   roleID,
		SenderName: roleName,
		Content:    content,
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return customerr.Wrap(err)
	}
	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdChatWorld), data)
	return customerr.Wrap(uc.publicGate.SendMessageAsync(ctx, "global", actorMsg))
}

func (uc *WorldChatUseCase) containsSensitive(content string) bool {
	words, ok := uc.configManager.GetSensitiveWordConfig()
	if !ok {
		return false
	}
	return chatdomain.ContainsSensitive(content, words)
}
