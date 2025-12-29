// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package dict_item

import (
	"context"
	"database/sql"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type DictItemCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictItemCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictItemCreateLogic {
	return &DictItemCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DictItemCreateLogic) DictItemCreate(req *types.DictItemCreateReq) error {
	if req == nil || req.TypeId == 0 || req.Label == "" || req.Value == "" {
		return errs.New(errs.CodeBadRequest, "字典类型ID、标签和值不能为空")
	}

	// 检查字典类型是否存在
	dictTypeRepo := repository.NewDictTypeRepository(l.svcCtx.Repository)
	_, err := dictTypeRepo.FindByID(l.ctx, req.TypeId)
	if err != nil {
		return errs.Wrap(errs.CodeBadRequest, "字典类型不存在", err)
	}

	sort := req.Sort
	if sort == 0 {
		sort = 0
	}
	status := req.Status
	if status == 0 {
		status = 1
	}

	dictItem := model.AdminDictItem{
		TypeId: req.TypeId,
		Label:  req.Label,
		Value:  req.Value,
		Sort:   sort,
		Status: status,
		Remark: sql.NullString{String: req.Remark, Valid: req.Remark != ""},
	}

	dictItemRepo := repository.NewDictItemRepository(l.svcCtx.Repository)
	if err := dictItemRepo.Create(l.ctx, &dictItem); err != nil {
		return errs.Wrap(errs.CodeInternalError, "创建字典项失败", err)
	}

	// 清除字典缓存
	cache := l.svcCtx.Repository.BusinessCache
	go func() {
		// 需要获取字典类型的 code 来清除缓存
		dictTypeRepo := repository.NewDictTypeRepository(l.svcCtx.Repository)
		dictType, err := dictTypeRepo.FindByID(context.Background(), req.TypeId)
		if err == nil {
			if err := cache.DeleteDictItems(context.Background(), dictType.Code); err != nil {
				l.Errorf("清除字典项缓存失败: code=%s, error=%v", dictType.Code, err)
			}
			if err := cache.DeleteDictItemsByType(context.Background(), req.TypeId); err != nil {
				l.Errorf("清除字典项缓存失败: typeId=%d, error=%v", req.TypeId, err)
			}
		}
	}()

	return nil
}
