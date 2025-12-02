package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	frienddomain "postapocgame/server/service/gameserver/internel/domain/friend"
	"postapocgame/server/service/gameserver/internel/domain/repository"
)

// FriendSystemAdapter 好友系统适配器
type FriendSystemAdapter struct {
	*BaseSystemAdapter
	playerRepo repository.PlayerRepository
}

// NewFriendSystemAdapter 创建好友系统适配器
func NewFriendSystemAdapter() *FriendSystemAdapter {
	return &FriendSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysFriend)),
		playerRepo:        di.GetContainer().PlayerGateway(),
	}
}

// OnInit 初始化好友数据
func (a *FriendSystemAdapter) OnInit(ctx context.Context) {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("friend sys OnInit get role err:%v", err)
		return
	}
	binaryData := playerRole.GetBinaryData()
	frienddomain.EnsureFriendData(binaryData)
}

// Ensure adapter implements ISystem
var _ iface.ISystem = (*FriendSystemAdapter)(nil)
