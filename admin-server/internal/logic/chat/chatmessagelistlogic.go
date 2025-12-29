// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"
	"strconv"
	"time"

	"postapocgame/admin-server/internal/model"
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

	// 获取当前用户
	user, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	messageRepo := repository.NewChatMessageRepository(l.svcCtx.Repository)
	userRepo := repository.NewUserRepository(l.svcCtx.Repository)

	// 如果查询私聊消息（userId > 0 且 roomId 为空），需要查询当前用户和指定用户之间的消息
	var list []model.ChatMessage
	var total int64
	if req.UserId > 0 && req.RoomId == "" {
		// 私聊：查询当前用户和指定用户之间的消息
		list, total, err = messageRepo.FindPrivateMessages(l.ctx, req.Page, req.PageSize, user.UserID, req.UserId)
	} else {
		// 群聊：查询房间消息
		list, total, err = messageRepo.FindPage(l.ctx, req.Page, req.PageSize, req.RoomId, 0)
	}
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

		// 查询接收用户信息
		toUserName := ""
		if msg.ToUserId > 0 {
			toUser, _ := userRepo.FindByID(l.ctx, msg.ToUserId)
			if toUser != nil {
				toUserName = toUser.Username
			}
		}

		items = append(items, types.ChatMessageItem{
			Id:           msg.Id,
			FromUserId:   msg.FromUserId,
			FromUserName: fromUserName,
			ToUserId:     msg.ToUserId,
			ToUserName:   toUserName,
			RoomId:       msg.RoomId,
			Content:      msg.Content,
			MessageType:  msg.MessageType,
			Status:       msg.Status,
			CreatedAt:    time.Unix(msg.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		})
	}

	return &types.ChatMessageListResp{
		Total: total,
		List:  items,
	}, nil
}
