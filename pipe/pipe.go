package pipe

import (
	"net"
)

type Pipe interface {
	Addr(name string) string
	Listen(addr string) (net.Listener, error)
	Dial(addr string) (net.Conn, error)
}

type Listener interface {
	Close() error
}
