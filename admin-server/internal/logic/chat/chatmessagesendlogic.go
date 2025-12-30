// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"
	"encoding/json"
	"time"

	"postapocgame/admin-server/internal/hub"
	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatMessageSendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatMessageSendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatMessageSendLogic {
	return &ChatMessageSendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatMessageSendLogic) ChatMessageSend(req *types.ChatMessageSendReq) (resp *types.ChatMessageSendResp, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	// 获取当前用户
	user, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	// 验证参数
	if req.Content == "" {
		return nil, errs.New(errs.CodeBadRequest, "消息内容不能为空")
	}
	if req.ChatId == 0 {
		return nil, errs.New(errs.CodeBadRequest, "聊天ID不能为空")
	}

	// 验证 chat 是否存在且用户有权限
	chatRepo := repository.NewChatRepository(l.svcCtx.Repository)
	chat, err := chatRepo.FindByID(l.ctx, req.ChatId)
	if err != nil {
		return nil, errs.Wrap(errs.CodeBadRequest, "聊天不存在", err)
	}
	if chat.DeletedAt != 0 {
		return nil, errs.New(errs.CodeBadRequest, "聊天已删除")
	}

	// 验证用户是否在聊天中
	chatUserRepo := repository.NewChatUserRepository(l.svcCtx.Repository)
	chatUsers, err := chatUserRepo.FindByChatID(l.ctx, req.ChatId)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询聊天成员失败", err)
	}
	hasPermission := false
	for _, cu := range chatUsers {
		if cu.UserId == user.UserID {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, errs.New(errs.CodeForbidden, "您不在该聊天中")
	}

	// 设置默认值
	if req.MessageType == 0 {
		req.MessageType = 1 // 默认文本消息
	}

	// 创建消息
	now := time.Now().Unix()
	message := &model.ChatMessage{
		ChatId:      req.ChatId,
		FromUserId:  user.UserID,
		Content:     req.Content,
		MessageType: req.MessageType,
		Status:      1, // 1已发送
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   0,
	}

	messageRepo := repository.NewChatMessageRepository(l.svcCtx.Repository)
	err = messageRepo.Create(l.ctx, message)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "发送消息失败", err)
	}

	// 通过 WebSocket 推送消息给聊天中的所有用户
	if l.svcCtx.ChatHub != nil {
		// 获取聊天中的所有用户ID
		userIDs := make([]uint64, 0, len(chatUsers))
		for _, cu := range chatUsers {
			userIDs = append(userIDs, cu.UserId)
		}

		chatMsg := &hub.ChatMessage{
			Type:      "chat", // 使用 "chat" 作为聊天消息类型
			FromID:    user.UserID,
			FromName:  user.Username,
			ChatID:    req.ChatId,
			Content:   req.Content,
			MessageID: message.Id,
			CreatedAt: message.CreatedAt,
		}
		messageBytes, err := json.Marshal(chatMsg)
		if err == nil {
			l.svcCtx.ChatHub.BroadcastToChat(req.ChatId, userIDs, messageBytes)
		} else {
			logx.Errorf("WebSocket 消息序列化失败: %v", err)
		}
	}

	return &types.ChatMessageSendResp{
		Id: message.Id,
	}, nil
}
