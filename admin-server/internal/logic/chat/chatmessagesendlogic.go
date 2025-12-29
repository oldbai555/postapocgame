// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"
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
	if req.RoomId == "" {
		return nil, errs.New(errs.CodeBadRequest, "聊天室ID不能为空")
	}

	// 设置默认值
	if req.MessageType == 0 {
		req.MessageType = 1 // 默认文本消息
	}

	// 创建消息
	now := time.Now().Unix()
	message := &model.ChatMessage{
		FromUserId:  user.UserID,
		ToUserId:    req.ToUserId,
		RoomId:      req.RoomId,
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

	// 通过 WebSocket 推送消息给接收用户
	if l.svcCtx.ChatHub != nil {
		chatMsg := &hub.ChatMessage{
			Type:      "chat", // 使用 "chat" 作为聊天消息类型
			FromID:    user.UserID,
			FromName:  user.Username,
			ToID:      req.ToUserId,
			RoomID:    req.RoomId,
			Content:   req.Content,
			MessageID: message.Id,
			CreatedAt: time.Unix(message.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		}
		if err := l.svcCtx.ChatHub.BroadcastChatMessage(chatMsg); err != nil {
			logx.Errorf("WebSocket 推送消息失败: %v", err)
			// 不返回错误，因为消息已经保存到数据库
		}
	}

	return &types.ChatMessageSendResp{
		Id: message.Id,
	}, nil
}
