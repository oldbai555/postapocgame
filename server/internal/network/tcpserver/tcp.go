/**
 * @Author: zjj
 * @Date: 2025/11/5
 * @Desc:
**/

package tcpserver

type TCPServer struct {
	Options
}

func NewTCPServer(opts ...Option) *TCPServer {
	opt := Options{}
	for _, option := range opts {
		option(&opt)
	}
	return &TCPServer{
		Options: opt,
	}
}

func (s *TCPServer) Start() error {
	return nil
}

func (s *TCPServer) Addr() string {
	return s.ServiceAddr
}

func (s *TCPServer) Stop() error {
	return nil
}
