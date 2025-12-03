package system

import (
	"context"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
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
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Warnf("AttrCalculator not found: RoleID=%d, SysID=%d, error: %v", roleID, sysID, err)
		return nil
	}
	attrCalcRaw := playerRole.GetAttrCalculator()
	attrCalc, ok := attrCalcRaw.(interfaces.IAttrCalculator)
	if !ok || attrCalc == nil {
		log.Warnf("AttrCalculator not found: RoleID=%d, SysID=%d", roleID, sysID)
		return nil
	}
	attrCalc.MarkDirty(sysID)
	return nil
}
