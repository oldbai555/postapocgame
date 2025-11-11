/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package entity

import (
	"postapocgame/server/internal/event"
)

func (pr *PlayerRole) Publish(typ event.Type, args ...interface{}) error {
	return nil
}
