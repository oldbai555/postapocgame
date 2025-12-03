package publicactor

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
)

var _ actor.IActorHandler = (*PublicHandler)(nil)

// NewPublicHandler 创建公共消息处理器
func NewPublicHandler() *PublicHandler {
	publicRole := NewPublicRole()
	handler := &PublicHandler{
		BaseActorHandler: actor.NewBaseActorHandler("public actor handler"),
		publicRole:       publicRole,
	}
	// 注意：RegisterMessageHandlers 需要在 PublicActor 创建后调用
	// 这里先不注册，等 PublicActor 创建后再注册
	// 加载数据
	handler.LoadData()
	return handler
}

// LoadData 从数据库加载数据
func (h *PublicHandler) LoadData() {
	// 加载离线数据
	h.publicRole.LoadOfflineData(context.Background())
}

// PublicHandler 公共消息处理器
type PublicHandler struct {
	*actor.BaseActorHandler
	actorCtx   actor.IActorContext // 存储 Actor Context 引用
	publicRole *PublicRole         // 公共角色数据
}

// SetActorContext 设置 Actor Context 引用
func (h *PublicHandler) SetActorContext(ctx actor.IActorContext) {
	h.actorCtx = ctx
}

func (h *PublicHandler) Loop() {
	// Loop 方法只负责触发 RunOne，不注册处理器
	if h.actorCtx != nil {
		ctx := context.Background()
		h.actorCtx.ExecuteAsync(actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgRunOne), []byte{}))
	}
}

// OnStart Actor 启动时调用（只调用一次）
func (h *PublicHandler) OnStart() {
	// 注册 RunOne 消息处理器
	h.RegisterMessageHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgRunOne), func(msg actor.IActorMessage) {
		ctx := msg.GetContext()
		handleRunOneMsg(ctx, msg, h.publicRole)
	})

	// 触发第一次 RunOne
	if h.actorCtx != nil {
		ctx := context.Background()
		h.actorCtx.ExecuteAsync(actor.NewBaseMessage(ctx, uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdDoRunOneMsg), []byte{}))
	}
}

// GetPublicRole 获取公共角色
func (h *PublicHandler) GetPublicRole() *PublicRole {
	return h.publicRole
}
