package system

import (
	"context"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// attrUseCaseAdapter implements interfaces.AttrUseCase for other systems.
type attrUseCaseAdapter struct{}

// NewAttrUseCaseAdapter creates an AttrUseCase adapter.
func NewAttrUseCaseAdapter() interfaces.AttrUseCase {
	return &attrUseCaseAdapter{}
}

// MarkDirty 标记需要重算的系统
func (a *attrUseCaseAdapter) MarkDirty(ctx context.Context, roleID uint64, sysID uint32) error {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Warnf("AttrCalculator not found: RoleID=%d, SysID=%d, error: %v", roleID, sysID, err)
		return customerr.Wrap(err)
	}
	attrCalc := playerRole.GetAttrCalculator()
	if attrCalc == nil {
		log.Warnf("AttrCalculator not found: RoleID=%d, SysID=%d", roleID, sysID)
		return nil
	}
	attrCalc.MarkDirty(sysID)
	return nil
}
