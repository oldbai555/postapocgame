/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package entity

import (
	"context"
	"postapocgame/server/internal/event"
)

func (pr *PlayerRole) Publish(typ event.Type, args ...interface{}) error {
	return pr.eventBus.Publish(context.Background(), &event.Event{
		Type: typ,
		Data: append([]interface{}{}, args...),
	})
}
