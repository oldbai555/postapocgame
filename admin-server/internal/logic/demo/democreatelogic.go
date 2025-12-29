// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package demo

import (
	"context"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type DemoCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDemoCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DemoCreateLogic {
	return &DemoCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DemoCreateLogic) DemoCreate(req *types.DemoCreateReq) error {
	if req == nil || req.Name == "" {
		return errs.New(errs.CodeBadRequest, "描述不能为空")
	}

	status := req.Status
	if status == 0 {
		status = 1
	}

	demo := model.Demo{
		Name:   req.Name,
		Status: status,
	}

	demoRepo := repository.NewDemoRepository(l.svcCtx.Repository)
	if err := demoRepo.Create(l.ctx, &demo); err != nil {
		return errs.Wrap(errs.CodeInternalError, "创建演示功能失败", err)
	}
	return nil
}
