package fuben

import (
	"errors"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/iface"
	"time"
)

var (
	getTimedFuBenForPlayer    func(sessionId string) (iface.IFuBen, bool)
	createTimedFuBenForPlayer func(sessionId string, name string, maxDuration time.Duration) (*FuBenSt, error)
)

// ErrTimedFuBenProviderMissing provider 未注册
var ErrTimedFuBenProviderMissing = errors.New("timed fuben provider not registered")

// RegisterTimedFuBenProvider 注册限时副本相关回调
func RegisterTimedFuBenProvider(getFunc func(string) (iface.IFuBen, bool), createFunc func(string, string, time.Duration) (*FuBenSt, error)) {
	getTimedFuBenForPlayer = getFunc
	createTimedFuBenForPlayer = createFunc
}

func getTimedFuBen(sessionId string) (iface.IFuBen, bool) {
	if getTimedFuBenForPlayer == nil {
		return nil, false
	}
	return getTimedFuBenForPlayer(sessionId)
}

func createTimedFuBen(sessionId string, name string, maxDuration time.Duration) (*FuBenSt, error) {
	if createTimedFuBenForPlayer == nil {
		return nil, ErrTimedFuBenProviderMissing
	}
	return createTimedFuBenForPlayer(sessionId, name, maxDuration)
}
