/**
 * @Author: zjj
 * @Date: 2025/11/5
 * @Desc:
**/

package tcpserver

type Options struct {
	Name        string // 服务器名称
	ServiceAddr string // current server service address (RPC)
}

type Option func(*Options)
