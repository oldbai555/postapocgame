package client

import (
	"fmt"
	"time"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
)

type waiter[T any] struct {
	ch chan T
}

func newWaiter[T any]() waiter[T] {
	return waiter[T]{ch: make(chan T, 1)}
}

func (w waiter[T]) Deliver(resp T) {
	select {
	case w.ch <- resp:
	default:
	}
}

func (w waiter[T]) Wait(timeout time.Duration) (T, error) {
	var zero T
	deadline := servertime.Now().Add(timeout)
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()
	select {
	case resp := <-w.ch:
		return resp, nil
	case <-timer.C:
		return zero, fmt.Errorf("wait response timeout (%s)", timeout)
	}
}

func (w waiter[T]) Chan() chan T {
	return w.ch
}

type flowRegistry struct {
	register   waiter[*protocol.S2CRegisterReq]
	login      waiter[*protocol.S2CLoginReq]
	roleList   waiter[*protocol.S2CRoleListReq]
	createRole waiter[*protocol.S2CCreateRoleReq]
	loginRole  waiter[*protocol.S2CLoginRoleReq]
	enterScene waiter[*protocol.S2CEnterSceneReq]
}

func newFlowRegistry() flowRegistry {
	return flowRegistry{
		register:   newWaiter[*protocol.S2CRegisterReq](),
		login:      newWaiter[*protocol.S2CLoginReq](),
		roleList:   newWaiter[*protocol.S2CRoleListReq](),
		createRole: newWaiter[*protocol.S2CCreateRoleReq](),
		loginRole:  newWaiter[*protocol.S2CLoginRoleReq](),
		enterScene: newWaiter[*protocol.S2CEnterSceneReq](),
	}
}
