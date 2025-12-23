package controller

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/playeractor/gateway"
	"postapocgame/server/service/gameserver/internel/playeractor/presenter"
	"postapocgame/server/service/gameserver/internel/playeractor/router"
	"postapocgame/server/service/gameserver/internel/playeractor/service/playerauth"

	"google.golang.org/protobuf/proto"

	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
)

// PlayerAccountController 负责账号注册/登录协议
type PlayerAccountController struct {
	registerUC    *playerauth.RegisterUseCase
	loginUC       *playerauth.LoginUseCase
	presenter     *presenter.PlayerAuthPresenter
	clientGateway gateway.ClientGateway
}

// NewPlayerAccountController 创建控制器
func NewPlayerAccountController() *PlayerAccountController {
	return &PlayerAccountController{
		registerUC:    playerauth.NewRegisterUseCase(deps.NewAccountRepository(), deps.NewTokenGenerator()),
		loginUC:       playerauth.NewLoginUseCase(deps.NewAccountRepository(), deps.NewTokenGenerator()),
		presenter:     presenter.NewPlayerAuthPresenter(deps.NewNetworkGateway()),
		clientGateway: deps.NewNetworkGateway(),
	}
}

// HandleRegister 处理注册协议
func (c *PlayerAccountController) HandleRegister(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := getSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SRegisterReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	result, err := c.registerUC.Execute(ctx, playerauth.RegisterInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	if result.Success {
		c.updateSessionAccount(sessionID, result.AccountID, result.Token)
	}

	return c.presenter.S2CRegister(ctx, sessionID, result)
}

// HandleLogin 处理登录协议
func (c *PlayerAccountController) HandleLogin(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := getSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SLoginReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	result, err := c.loginUC.Execute(ctx, playerauth.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	if result.Success {
		c.updateSessionAccount(sessionID, result.AccountID, result.Token)
	}

	return c.presenter.S2CLogin(ctx, sessionID, result)
}

// HandleVerify 暂未实现
func (c *PlayerAccountController) HandleVerify(context.Context, *network.ClientMessage) error {
	return nil
}

func (c *PlayerAccountController) updateSessionAccount(sessionID string, accountID uint64, token string) {
	session := c.clientGateway.GetSession(sessionID)
	if session == nil {
		return
	}
	session.SetAccountID(uint(accountID))
	session.SetToken(token)
}

func getSessionIDFromContext(ctx context.Context) (string, error) {
	sessionID, _ := ctx.Value(gshare.ContextKeySession).(string)
	if sessionID == "" {
		return "", customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session not found in context")
	}
	return sessionID, nil
}
func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		accountController := NewPlayerAccountController()
		router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SRegister), accountController.HandleRegister)
		router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SLogin), accountController.HandleLogin)
		router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SVerify), accountController.HandleVerify)
	})
}
