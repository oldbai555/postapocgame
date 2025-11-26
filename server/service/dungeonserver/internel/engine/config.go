/**
 * @Author: zjj
 * @Date: 2025/11/6
 * @Desc:
**/

package engine

import (
	"fmt"
	"net"
	"os"
	"path"
	"postapocgame/server/internal"
	"postapocgame/server/pkg/tool"
)

// ServerConfig DungeonServer配置
type ServerConfig struct {
	SrvType uint8  `json:"srv_type"` // 服务类型: 1=默认副本服务器
	TCPAddr string `json:"tcp_addr"` // TCP监听地址

	// Actor配置
	ActorMailboxSize int `json:"actor_mailbox_size"` // Actor邮箱大小
}

const (
	defaultSrvType        = 1
	defaultActorMailboxDS = 1024
)

func (c *ServerConfig) applyDefaults() {
	if c.SrvType == 0 {
		c.SrvType = defaultSrvType
	}
	if c.ActorMailboxSize <= 0 {
		c.ActorMailboxSize = defaultActorMailboxDS
	}
}

func (c *ServerConfig) Validate() error {
	if c.TCPAddr == "" {
		return fmt.Errorf("tcp_addr is required")
	}
	if err := validateAddr(c.TCPAddr); err != nil {
		return fmt.Errorf("invalid tcp_addr: %w", err)
	}
	if c.ActorMailboxSize <= 0 {
		return fmt.Errorf("actor_mailbox_size must be greater than 0")
	}
	return nil
}

func validateAddr(addr string) error {
	if addr == "" {
		return fmt.Errorf("address is empty")
	}
	_, _, err := net.SplitHostPort(addr)
	return err
}

func LoadServerConfig(confPath string) (*ServerConfig, error) {
	if confPath == "" {
		confPath = path.Join(tool.GetCurDir(), "dungeonsrv.json")
	}
	bytes, err := os.ReadFile(confPath)
	if err != nil {
		return nil, err
	}
	var conf ServerConfig
	err = internal.Unmarshal(bytes, &conf)
	if err != nil {
		return nil, err
	}

	conf.applyDefaults()
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	return &conf, nil
}
