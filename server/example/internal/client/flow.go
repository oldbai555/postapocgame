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
	register     waiter[*protocol.S2CRegisterResultReq]
	login        waiter[*protocol.S2CLoginResultReq]
	roleList     waiter[*protocol.S2CRoleListReq]
	createRole   waiter[*protocol.S2CCreateRoleResultReq]
	enterScene   waiter[*protocol.S2CEnterSceneReq]
	aoi          waiter[*EntityView]
	skillDamage  waiter[*protocol.S2CSkillDamageResultReq]
	bagData      waiter[*protocol.S2CBagDataReq]
	moneyData    waiter[*protocol.S2CMoneyDataReq]
	gmResult     waiter[*protocol.S2CGMCommandResultReq]
	useItem      waiter[*protocol.S2CUseItemResultReq]
	pickup       waiter[*protocol.S2CPickupItemResultReq]
	dungeonEnter waiter[*protocol.S2CEnterDungeonResultReq]
}

func newFlowRegistry() flowRegistry {
	return flowRegistry{
		register:     newWaiter[*protocol.S2CRegisterResultReq](),
		login:        newWaiter[*protocol.S2CLoginResultReq](),
		roleList:     newWaiter[*protocol.S2CRoleListReq](),
		createRole:   newWaiter[*protocol.S2CCreateRoleResultReq](),
		enterScene:   newWaiter[*protocol.S2CEnterSceneReq](),
		aoi:          newWaiter[*EntityView](),
		skillDamage:  newWaiter[*protocol.S2CSkillDamageResultReq](),
		bagData:      newWaiter[*protocol.S2CBagDataReq](),
		moneyData:    newWaiter[*protocol.S2CMoneyDataReq](),
		gmResult:     newWaiter[*protocol.S2CGMCommandResultReq](),
		useItem:      newWaiter[*protocol.S2CUseItemResultReq](),
		pickup:       newWaiter[*protocol.S2CPickupItemResultReq](),
		dungeonEnter: newWaiter[*protocol.S2CEnterDungeonResultReq](),
	}
}
