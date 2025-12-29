// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package demo

import (
	"context"

	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type DemoUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDemoUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DemoUpdateLogic {
	return &DemoUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DemoUpdateLogic) DemoUpdate(req *types.DemoUpdateReq) error {
	if req == nil || req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "请求参数不能为空")
	}

	demoRepo := repository.NewDemoRepository(l.svcCtx.Repository)
	demo, err := demoRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return errs.Wrap(errs.CodeNotFound, "演示功能不存在", err)
	}

	// 更新字段
	if req.Name != "" {
		demo.Name = req.Name
	}
	if req.Status == 0 || req.Status == 1 {
		demo.Status = req.Status
	}

	if err := demoRepo.Update(l.ctx, demo); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新演示功能失败", err)
	}
	return nil
}
