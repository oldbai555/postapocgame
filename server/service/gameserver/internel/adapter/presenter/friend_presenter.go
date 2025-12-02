package presenter

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/adapter/gateway"
)

// FriendPresenter 好友系统响应构建器
type FriendPresenter struct {
	network gateway.NetworkGateway
}

// NewFriendPresenter 创建 Presenter
func NewFriendPresenter(network gateway.NetworkGateway) *FriendPresenter {
	return &FriendPresenter{network: network}
}

func (p *FriendPresenter) SendAddFriendResult(ctx context.Context, sessionID string, success bool, message string, targetID uint64) error {
	resp := &protocol.S2CAddFriendResultReq{
		Success:  success,
		Message:  message,
		TargetId: targetID,
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CAddFriendResult), resp)
}

func (p *FriendPresenter) SendRespondResult(ctx context.Context, sessionID string, success bool, message string, requesterID uint64, accepted bool) error {
	resp := &protocol.S2CRespondFriendReqResultReq{
		Success:     success,
		Message:     message,
		RequesterId: requesterID,
		Accepted:    accepted,
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CRespondFriendReqResult), resp)
}

func (p *FriendPresenter) SendRemoveResult(ctx context.Context, sessionID string, success bool, message string, friendID uint64) error {
	resp := &protocol.S2CRemoveFriendResultReq{
		Success:  success,
		Message:  message,
		FriendId: friendID,
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CRemoveFriendResult), resp)
}

func (p *FriendPresenter) SendAddBlacklistResult(ctx context.Context, sessionID string, success bool, message string) error {
	resp := &protocol.S2CAddToBlacklistResultReq{
		Success: success,
		Message: message,
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CAddToBlacklistResult), resp)
}

func (p *FriendPresenter) SendRemoveBlacklistResult(ctx context.Context, sessionID string, success bool, message string) error {
	resp := &protocol.S2CRemoveFromBlacklistResultReq{
		Success: success,
		Message: message,
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CRemoveFromBlacklistResult), resp)
}

func (p *FriendPresenter) SendBlacklist(ctx context.Context, sessionID string, ids []uint64) error {
	resp := &protocol.S2CBlacklistReq{
		BlacklistIds: ids,
	}
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CBlacklist), resp)
}

// SendError 通用错误
func (p *FriendPresenter) SendError(ctx context.Context, sessionID string, message string) error {
	return p.network.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
		Code: -1,
		Msg:  message,
	})
}
