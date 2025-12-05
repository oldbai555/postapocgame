/**
 * @Author: zjj
 * @Date: 2025/11/6
 * @Desc:
**/

package engine

import (
	"net"
	"os"
	"path"
	"postapocgame/server/internal"
	"postapocgame/server/internal/actor"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/tool"
)

// ServerConfig GameServer配置
type ServerConfig struct {
	AppID      uint32 `json:"app_id"`      // 应用ID
	PlatformID uint32 `json:"platform_id"` // 平台ID
	SrvId      uint32 `json:"srv_id"`      // 区服ID

	// TCP配置
	TCPAddr         string   `json:"tcp_addr"`          // TCP监听地址
	GatewayAllowIPs []string `json:"gateway_allow_ips"` // 允许连接的网关IP列表

	// Actor配置
	ActorMode        actor.ActorMode `json:"actor_mode"`         // Actor类型: 0=Single, 1=PerPlayer
	ActorPoolSize    int             `json:"actor_pool_size"`    // Actor池大小
	ActorMailboxSize int             `json:"actor_mailbox_size"` // Actor邮箱大小

	// DungeonServer配置
	DungeonServerAddrMap map[uint8]string `json:"dungeon_server_addr_map"` // DungeonServer地址映射 [srvType]addr
}

const (
	defaultActorMailboxSize = 1024
	defaultActorPoolSize    = 1
)

func (c *ServerConfig) applyDefaults() {
	if c.ActorMailboxSize <= 0 {
		c.ActorMailboxSize = defaultActorMailboxSize
	}
	if c.ActorPoolSize <= 0 {
		c.ActorPoolSize = defaultActorPoolSize
	}
	if c.ActorMode != actor.ModeSingle && c.ActorMode != actor.ModePerKey {
		c.ActorMode = actor.ModePerKey
	}
}

func (c *ServerConfig) Validate() error {
	if c.AppID == 0 {
		return customerr.NewError("app_id must be greater than 0")
	}
	if c.PlatformID == 0 {
		return customerr.NewError("platform_id must be greater than 0")
	}
	if c.SrvId == 0 {
		return customerr.NewError("srv_id must be greater than 0")
	}
	if c.ActorMailboxSize <= 0 {
		return customerr.NewError("actor_mailbox_size must be greater than 0")
	}
	if c.ActorPoolSize <= 0 {
		return customerr.NewError("actor_pool_size must be greater than 0")
	}
	// InProcess DungeonActor 模式下，DungeonServerAddrMap 可为空；
	// 如需远程 DungeonServer，可在配置中补充并复用现有校验逻辑。
	if len(c.DungeonServerAddrMap) > 0 {
		for srvType, addr := range c.DungeonServerAddrMap {
			if addr == "" {
				return customerr.NewError("dungeon_server_addr_map[%d] is empty", srvType)
			}
			if err := validateAddr(addr); err != nil {
				return customerr.NewError("invalid dungeon server addr for srvType=%d: %v", srvType, err)
			}
		}
	}
	if c.TCPAddr != "" {
		if err := validateAddr(c.TCPAddr); err != nil {
			return customerr.NewError("invalid tcp_addr: %v", err)
		}
	}
	return nil
}

func validateAddr(addr string) error {
	if addr == "" {
		return customerr.NewError("address is empty")
	}
	_, _, err := net.SplitHostPort(addr)
	return err
}

func LoadServerConfig(confPath string) (*ServerConfig, error) {
	if confPath == "" {
		confPath = path.Join(tool.GetCurDir(), "gamesrv.json")
	}
	bytes, err := os.ReadFile(confPath)
	if err != nil {
		return nil, customerr.Wrap(err)
	}
	var conf ServerConfig
	err = internal.Unmarshal(bytes, &conf)
	if err != nil {
		return nil, customerr.Wrap(err)
	}

	conf.applyDefaults()
	if err := conf.Validate(); err != nil {
		return nil, customerr.Wrap(err)
	}

	return &conf, nil
}

// 热加载配置建议示例（可在main里用go热重载）
// func (c *ServerConfig) WatchAndReload() {
//     go func() {
//         for {
//             // 监听文件变化或定时reload
//             <-time.After(30 * time.Second)
//             // TODO: reload from disk
//             // ...变更后赋值到c...
//         }
//     }()
// }
