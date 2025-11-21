/**
 * @Author: zjj
 * @Date: 2025/11/5
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
	"time"
)

// Config Gateway配置
type Config struct {
	// 游戏服务器地址
	GameServerAddr string `json:"gameServerAddr"`

	// TCP配置
	TCPAddr string `json:"tcp_addr"` // TCP监听地址,如 ":8080"

	// WebSocket配置
	WSAddr string `json:"ws_addr"` // WebSocket监听地址,如 ":8081"
	WSPath string `json:"ws_path"` // WebSocket路径,如 "/ws"

	// 会话配置
	SessionBufferSize int           // 每个会话的发送缓冲区大小
	MaxSessions       uint32        // 最大会话数
	SessionTimeout    time.Duration // 会话超时时间

	MaxFrameSize int // 帧协议配置 最大帧大小
}

const (
	defaultWSPath            = "/ws"
	defaultSessionBufferSize = 256
	defaultMaxSessions       = 10000
	defaultSessionTimeout    = 5 * time.Minute
	defaultMaxFrameSize      = 10 * 1024 * 1024 // 10MB
)

func (c *Config) applyDefaults() {
	if c.WSPath == "" {
		c.WSPath = defaultWSPath
	}
	if c.SessionBufferSize <= 0 {
		c.SessionBufferSize = defaultSessionBufferSize
	}
	if c.MaxSessions == 0 {
		c.MaxSessions = defaultMaxSessions
	}
	if c.SessionTimeout <= 0 {
		c.SessionTimeout = defaultSessionTimeout
	}
	if c.MaxFrameSize <= 0 {
		c.MaxFrameSize = defaultMaxFrameSize
	}
}

func (c *Config) Validate() error {
	if c.GameServerAddr == "" {
		return fmt.Errorf("gameServerAddr is required")
	}
	if err := validateAddr(c.GameServerAddr); err != nil {
		return fmt.Errorf("invalid gameServerAddr: %w", err)
	}
	if c.TCPAddr == "" && c.WSAddr == "" {
		return fmt.Errorf("at least one of tcp_addr or ws_addr must be configured")
	}
	if c.TCPAddr != "" {
		if err := validateAddr(c.TCPAddr); err != nil {
			return fmt.Errorf("invalid tcp_addr: %w", err)
		}
	}
	if c.WSAddr != "" {
		if err := validateAddr(c.WSAddr); err != nil {
			return fmt.Errorf("invalid ws_addr: %w", err)
		}
	}
	if c.SessionBufferSize <= 0 {
		return fmt.Errorf("sessionBufferSize must be greater than 0")
	}
	if c.MaxSessions == 0 {
		return fmt.Errorf("maxSessions must be greater than 0")
	}
	if c.SessionTimeout <= 0 {
		return fmt.Errorf("sessionTimeout must be greater than 0")
	}
	if c.MaxFrameSize <= 0 {
		return fmt.Errorf("maxFrameSize must be greater than 0")
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

func LoadGatewayConf(confPath string) (*Config, error) {
	if confPath == "" {
		confPath = path.Join(tool.GetCurDir(), "gateway.json")
	}
	bytes, err := os.ReadFile(confPath)
	if err != nil {
		return nil, err
	}
	var conf Config
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
