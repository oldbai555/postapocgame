package timesync

import (
	"context"
	"sync"
	"time"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/service/gameserver/internel/manager"
)

const defaultInterval = time.Second

type Broadcaster struct {
	interval time.Duration
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewBroadcaster(interval time.Duration) *Broadcaster {
	if interval <= 0 {
		interval = defaultInterval
	}
	return &Broadcaster{interval: interval}
}

func (b *Broadcaster) Start(ctx context.Context) {
	if b == nil {
		return
	}
	innerCtx, cancel := context.WithCancel(ctx)
	b.cancel = cancel
	b.wg.Add(1)
	go b.loop(innerCtx)
}

func (b *Broadcaster) Stop() {
	if b == nil {
		return
	}
	if b.cancel != nil {
		b.cancel()
	}
	b.wg.Wait()
}

func (b *Broadcaster) loop(ctx context.Context) {
	defer b.wg.Done()
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.broadcast()
		}
	}
}

func (b *Broadcaster) broadcast() {
	roles := manager.GetPlayerRoleManager().GetAll()
	if len(roles) == 0 {
		return
	}
	resp := &protocol.S2CTimeSyncReq{
		ServerTimeMs: servertime.UnixMilli(),
	}
	for _, role := range roles {
		if role == nil {
			continue
		}
		_ = role.SendProtoMessage(uint16(protocol.S2CProtocol_S2CTimeSync), resp)
	}
}
