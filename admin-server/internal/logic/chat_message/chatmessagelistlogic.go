// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_message

import (
	"context"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatMessageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatMessageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatMessageListLogic {
	return &ChatMessageListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatMessageListLogic) ChatMessageList(req *types.ChatMessageListReq) (resp *types.ChatMessageListResp, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20 // 管理页面默认20条
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // 限制最大100条
	}

	// 获取当前用户（用于权限验证）
	_, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	messageRepo := repository.NewChatMessageRepository(l.svcCtx.Repository)
	userRepo := repository.NewUserRepository(l.svcCtx.Repository)

	// 根据 chatId 查询消息，如果 chatId == 0，则查询所有消息（管理页面）
	list, total, err := messageRepo.FindByChatID(l.ctx, req.Page, req.PageSize, req.ChatId)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询聊天消息列表失败", err)
	}

	items := make([]types.ChatMessageItem, 0, len(list))
	for _, msg := range list {
		// 查询发送用户信息
		fromUser, _ := userRepo.FindByID(l.ctx, msg.FromUserId)
		fromUserName := ""
		if fromUser != nil {
			fromUserName = fromUser.Username
		}

		items = append(items, types.ChatMessageItem{
			Id:           msg.Id,
			ChatId:       msg.ChatId,
			FromUserId:   msg.FromUserId,
			FromUserName: fromUserName,
			Content:      msg.Content,
			MessageType:  msg.MessageType,
			Status:       msg.Status,
			CreatedAt:    msg.CreatedAt,
		})
	}

	return &types.ChatMessageListResp{
		Total: total,
		List:  items,
	}, nil
}
