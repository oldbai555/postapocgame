package user

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserListLogic {
	return &UserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserListLogic) UserList(req *types.UserListReq) (resp *types.UserListResp, err error) {
	if req == nil {
		return nil, errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	userRepo := repository.NewUserRepository(l.svcCtx.Repository)
	list, total, err := userRepo.FindPage(l.ctx, req.Page, req.PageSize, req.Username)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInternalError, "查询用户列表失败", err)
	}

	items := make([]types.UserItem, 0, len(list))
	for _, u := range list {
		items = append(items, types.UserItem{
			Id:           u.Id,
			Username:     u.Username,
			Nickname:     u.Nickname,
			Avatar:       u.Avatar,
			Signature:    u.Signature,
			DepartmentId: u.DepartmentId,
			Status:       u.Status,
			CreatedAt:    int64(u.CreatedAt),
		})
	}

	return &types.UserListResp{
		Total: total,
		List:  items,
	}, nil
}
