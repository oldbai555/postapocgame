package system

import (
	"context"
	"postapocgame/server/service/gameserver/internel/core/iface"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/mail"
)

// MailSystemAdapter 邮件系统适配器
//
// 生命周期职责：
// - OnInit: 调用 InitMailDataUseCase 初始化邮件数据（邮件列表结构）
// - 其他生命周期: 暂未使用
//
// 业务逻辑：所有业务逻辑（发邮件、收件、附件领取）均在 UseCase 层实现
//
// ⚠️ 防退化机制：禁止在 SystemAdapter 中编写业务规则逻辑，只允许调用 UseCase 与管理生命周期
type MailSystemAdapter struct {
	*BaseSystemAdapter
	initMailDataUseCase *mail.InitMailDataUseCase
}

func NewMailSystemAdapter() *MailSystemAdapter {
	container := di.GetContainer()
	initMailDataUC := mail.NewInitMailDataUseCase(container.PlayerGateway())
	return &MailSystemAdapter{
		BaseSystemAdapter:   NewBaseSystemAdapter(uint32(protocol.SystemId_SysMail)),
		initMailDataUseCase: initMailDataUC,
	}
}

// OnInit 初始化邮件数据
func (a *MailSystemAdapter) OnInit(ctx context.Context) {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("mail sys OnInit get role err:%v", err)
		return
	}
	// 初始化邮件数据（包括邮件列表结构等业务逻辑）
	if err := a.initMailDataUseCase.Execute(ctx, playerRole.GetPlayerRoleId()); err != nil {
		log.Errorf("mail sys OnInit init mail data err:%v", err)
		return
	}
}

var _ iface.ISystem = (*MailSystemAdapter)(nil)
