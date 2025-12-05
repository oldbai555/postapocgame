package chat

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	chatdomain "postapocgame/server/service/gameserver/internel/app/playeractor/domain/chat"
	interfaces2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
	"time"
)

const (
	worldChatMaxRunes = 200
	worldChatCooldown = 10 * time.Second
)

// WorldChatUseCase 世界聊天
type WorldChatUseCase struct {
	configManager interfaces2.ConfigManager
	publicGate    interfaces2.PublicActorGateway
}

func NewWorldChatUseCase(configManager interfaces2.ConfigManager, publicGate interfaces2.PublicActorGateway) *WorldChatUseCase {
	return &WorldChatUseCase{
		configManager: configManager,
		publicGate:    publicGate,
	}
}

func (uc *WorldChatUseCase) Execute(
	ctx context.Context,
	limiter interfaces2.ChatRateLimiter,
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
		return customerr.NewError("聊天内容验证失败: %s", reason)
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
	words := uc.configManager.GetSensitiveWordConfig()
	if len(words) == 0 {
		return false
	}
	return chatdomain.ContainsSensitive(content, words)
}
