// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	"strconv"

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

	// 从字典获取聊天窗口消息数量限制
	chatMessageLimit := int64(30) // 默认30
	dictTypeRepo := repository.NewDictTypeRepository(l.svcCtx.Repository)
	dictType, err := dictTypeRepo.FindByCode(l.ctx, "chat_config")
	if err == nil && dictType != nil {
		dictItemRepo := repository.NewDictItemRepository(l.svcCtx.Repository)
		items, err := dictItemRepo.FindByTypeID(l.ctx, dictType.Id)
		if err == nil {
			for _, item := range items {
				if item.Label == "聊天窗口消息数量" && item.Value != "" {
					if limit, parseErr := strconv.ParseInt(item.Value, 10, 64); parseErr == nil && limit > 0 {
						chatMessageLimit = limit
						break
					}
				}
			}
		}
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = chatMessageLimit // 使用从字典获取的限制值
	}
	// 限制最大页面大小不超过字典配置的值
	if req.PageSize > chatMessageLimit {
		req.PageSize = chatMessageLimit
	}

	messageRepo := repository.NewChatMessageRepository(l.svcCtx.Repository)
	userRepo := repository.NewUserRepository(l.svcCtx.Repository)

	// 根据 chatId 查询消息
	if req.ChatId == 0 {
		return nil, errs.New(errs.CodeBadRequest, "chatId 不能为空")
	}

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
