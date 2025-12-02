package system

import (
	"context"
	"postapocgame/server/service/gameserver/internel/core/iface"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	maildomain "postapocgame/server/service/gameserver/internel/domain/mail"
)

// MailSystemAdapter 邮件系统适配器
type MailSystemAdapter struct {
	*BaseSystemAdapter
}

func NewMailSystemAdapter() *MailSystemAdapter {
	return &MailSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysMail)),
	}
}

// OnInit 初始化邮件数据
func (a *MailSystemAdapter) OnInit(ctx context.Context) {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("mail sys OnInit get role err:%v", err)
		return
	}
	maildomain.EnsureMailData(playerRole.GetBinaryData())
}

var _ iface.ISystem = (*MailSystemAdapter)(nil)
