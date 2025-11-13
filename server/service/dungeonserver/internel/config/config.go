/**
 * @Author: zjj
 * @Date: 2025/11/6
 * @Desc:
**/

package config

import (
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

	return &conf, nil
}
