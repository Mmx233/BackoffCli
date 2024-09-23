//go:build !windows

package pipe

import (
	"fmt"
	"net"
	"os"
	"path"
)

func New() Pipe {
	return _Pipe{}
}

type _Pipe struct{}

func (_Pipe) Addr(name string) string {
	return fmt.Sprintf(`/tmp/%s`, name)
}

func (_Pipe) Listen(addr string) (net.Listener, error) {
	if err := os.MkdirAll(path.Dir(addr), 0777); err != nil {
		return nil, err
	}
	if err := os.RemoveAll(addr); err != nil {
		return nil, err
	}
	return net.Listen("unix", addr)
}

func (_Pipe) Dial(addr string) (net.Conn, error) {
	return net.Dial("unix", addr)
}
