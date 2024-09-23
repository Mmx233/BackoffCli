package pipe

import (
	"fmt"
	"github.com/natefinch/npipe"
	"net"
)

func New() Pipe {
	return _Pipe{}
}

type _Pipe struct{}

func (_Pipe) Addr(name string) string {
	return fmt.Sprintf(`\\.\pipe\%s`, name)
}

func (_Pipe) Listen(addr string) (net.Listener, error) {
	return npipe.Listen(addr)
}

func (_Pipe) Dial(addr string) (net.Conn, error) {
	return npipe.Dial(addr)
}
