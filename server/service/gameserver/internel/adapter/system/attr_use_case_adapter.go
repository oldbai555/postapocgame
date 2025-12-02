package system

import (
	"context"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// attrUseCaseAdapter implements interfaces.AttrUseCase for other systems.
type attrUseCaseAdapter struct{}

// NewAttrUseCaseAdapter creates an AttrUseCase adapter.
func NewAttrUseCaseAdapter() interfaces.AttrUseCase {
	return &attrUseCaseAdapter{}
}

// MarkDirty 标记需要重算的系统
func (a *attrUseCaseAdapter) MarkDirty(ctx context.Context, roleID uint64, sysID uint32) error {
	attrSys := GetAttrSys(ctx)
	if attrSys == nil {
		log.Warnf("AttrSys not found: RoleID=%d, SysID=%d", roleID, sysID)
		return nil
	}
	attrSys.MarkDirty(sysID)
	return nil
}
