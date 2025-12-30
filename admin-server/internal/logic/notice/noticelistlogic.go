// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package notice

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type NoticeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNoticeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NoticeListLogic {
	return &NoticeListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NoticeListLogic) NoticeList(req *types.NoticeListReq) (resp *types.NoticeListResp, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 处理筛选条件：如果未传入，使用 -1 作为标记
	noticeType := req.NoticeType
	if noticeType < 0 {
		noticeType = 0 // 未传入时设为0，表示不筛选类型
	}
	status := req.Status
	if status < 0 {
		status = -1 // 未传入时设为-1，表示不筛选状态
	}
	// 状态：1=草稿，2=已发布，0=未定义（不使用）

	noticeRepo := repository.NewNoticeRepository(l.svcCtx.Repository)
	list, total, err := noticeRepo.FindPage(l.ctx, req.Page, req.PageSize, req.Title, noticeType, status)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询公告列表失败", err)
	}

	items := make([]types.NoticeItem, 0, len(list))
	for _, n := range list {
		items = append(items, types.NoticeItem{
			Id:          n.Id,
			Title:       n.Title,
			Content:     n.Content,
			NoticeType:  n.Type,
			Status:      n.Status,
			PublishTime: n.PublishTime,
			CreatedBy:   n.CreatedBy,
			CreatedAt:   n.CreatedAt,
			UpdatedAt:   n.UpdatedAt,
		})
	}

	return &types.NoticeListResp{
		Total: total,
		List:  items,
	}, nil
}
