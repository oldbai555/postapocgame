// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_group

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatGroupListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatGroupListLogic {
	return &ChatGroupListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatGroupListLogic) ChatGroupList(req *types.ChatGroupListReq) (resp *types.ChatGroupListResp, err error) {
	// 获取当前用户
	_, ok := jwthelper.FromContext(l.ctx)
	if !ok {
		return nil, errs.New(errs.CodeUnauthorized, "未登录或登录已过期")
	}

	// 参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	chatRepo := repository.NewChatRepository(l.svcCtx.Repository)

	// 查询群组列表
	groups, total, err := chatRepo.FindGroups(l.ctx, req.Page, req.PageSize, req.Name)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询群组列表失败", err)
	}

	// 获取每个群组的成员数量
	items := make([]types.ChatGroupItem, 0, len(groups))
	for _, group := range groups {
		memberCount, err := chatRepo.CountMembersByChatID(l.ctx, group.Id)
		if err != nil {
			logx.Errorf("统计群组 %d 成员数量失败: %v", group.Id, err)
			memberCount = 0
		}

		items = append(items, types.ChatGroupItem{
			Id:          group.Id,
			Name:        group.Name,
			Avatar:      group.Avatar,
			Description: group.Description,
			CreatedBy:   group.CreatedBy,
			CreatedAt:   group.CreatedAt,
			MemberCount: memberCount,
		})
	}

	resp = &types.ChatGroupListResp{
		Total: total,
		List:  items,
	}

	return resp, nil
}
