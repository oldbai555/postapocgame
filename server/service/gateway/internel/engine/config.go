/**
 * @Author: zjj
 * @Date: 2025/11/5
 * @Desc:
**/

package engine

import (
	"os"
	"path"
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
	MaxSessions       int           // 最大会话数
	SessionTimeout    time.Duration // 会话超时时间

	MaxFrameSize int // 帧协议配置 最大帧大小
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
	err = tool.JsonUnmarshal(bytes, &conf)
	if err != nil {
		return nil, err
	}
	conf.WSPath = "/ws"
	conf.SessionBufferSize = 256
	conf.MaxSessions = 10000
	conf.SessionTimeout = 5 * time.Minute
	conf.MaxFrameSize = 10 * 1024 * 1024 // 10MB
	return &conf, nil
}
