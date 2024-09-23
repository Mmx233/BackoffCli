//go:build !windows

package pipe

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
)

func New() Pipe {
	return _Pipe{
		Dialer: &net.Dialer{},
	}
}

type _Pipe struct {
	Dialer *net.Dialer
}

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

func (p _Pipe) Dial(ctx context.Context, addr string) (net.Conn, error) {
	return p.Dialer.DialContext(ctx, "unix", addr)
}
