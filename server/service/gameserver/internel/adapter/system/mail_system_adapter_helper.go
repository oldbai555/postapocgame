package system

import (
	"context"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// GetMailSys 获取邮件系统适配器
func GetMailSys(ctx context.Context) *MailSystemAdapter {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	sys := playerRole.GetSystem(uint32(protocol.SystemId_SysMail))
	if sys == nil {
		return nil
	}
	mailSys, ok := sys.(*MailSystemAdapter)
	if !ok || !mailSys.IsOpened() {
		return nil
	}
	return mailSys
}
