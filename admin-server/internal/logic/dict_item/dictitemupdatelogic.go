// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package dict_item

import (
	"context"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type DictItemUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictItemUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictItemUpdateLogic {
	return &DictItemUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DictItemUpdateLogic) DictItemUpdate(req *types.DictItemUpdateReq) error {
	if req == nil || req.Id == 0 {
		return errs.New(errs.CodeBadRequest, "字典项ID不能为空")
	}

	dictItemRepo := repository.NewDictItemRepository(l.svcCtx.Repository)
	dictItem, err := dictItemRepo.FindByID(l.ctx, req.Id)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询字典项失败", err)
	}

	if req.Label != "" {
		dictItem.Label = req.Label
	}
	if req.Value != "" {
		dictItem.Value = req.Value
	}
	// Status 字段：0 是有效值（禁用），需要特殊处理
	// 由于无法判断字段是否提供，我们检查 status 是否在有效范围内（0 或 1）
	if req.Status == 0 || req.Status == 1 {
		dictItem.Status = req.Status
	}
	// Sort 字段：0 也是有效值，需要特殊处理
	// 由于 sort 通常 >= 0，我们检查是否 >= 0（假设 sort 不会是负数）
	if req.Sort >= 0 {
		dictItem.Sort = req.Sort
	}

	if err := dictItemRepo.Update(l.ctx, dictItem); err != nil {
		return errs.Wrap(errs.CodeInternalError, "更新字典项失败", err)
	}

	// 清除字典缓存
	cache := l.svcCtx.Repository.BusinessCache
	go func() {
		// 需要获取字典类型的 code 来清除缓存
		dictTypeRepo := repository.NewDictTypeRepository(l.svcCtx.Repository)
		dictType, err := dictTypeRepo.FindByID(context.Background(), dictItem.TypeId)
		if err == nil {
			if err := cache.DeleteDictItems(context.Background(), dictType.Code); err != nil {
				l.Errorf("清除字典项缓存失败: code=%s, error=%v", dictType.Code, err)
			}
			if err := cache.DeleteDictItemsByType(context.Background(), dictItem.TypeId); err != nil {
				l.Errorf("清除字典项缓存失败: typeId=%d, error=%v", dictItem.TypeId, err)
			}
		}
	}()

	return nil
}
