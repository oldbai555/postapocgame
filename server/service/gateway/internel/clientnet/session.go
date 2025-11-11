/**
 * @Author: zjj
 * @Date: 2025/11/10
 * @Desc:
**/

package clientnet

import (
	"net"
	"sync"
	"time"
)

// Session 统一的会话抽象
type Session struct {
	Id         string       // 会话ID
	Addr       net.Addr     // 客户端地址
	ConnType   ConnType     // 连接类型
	State      SessionState // 会话状态
	UserId     string       // 用户ID(认证后设置)
	SendChan   chan []byte  // 发送消息通道
	CreatedAt  time.Time    // 创建时间
	LastActive time.Time    // 最后活跃时间
	closeOnce  sync.Once
}

// SafeClose ensures SendChan is closed exactly once
func (s *Session) SafeClose() {
	s.closeOnce.Do(func() {
		// protect against nil
		if s.SendChan != nil {
			close(s.SendChan)
		}
	})
}
