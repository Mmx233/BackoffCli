package pipe

import (
	"context"
	"fmt"
	"github.com/Microsoft/go-winio"
	"net"
	"time"
)

func New() Pipe {
	return _Pipe{}
}

type _Pipe struct{}

func (_Pipe) Addr(name string) string {
	return fmt.Sprintf(`\\.\pipe\%s`, name)
}

func (_Pipe) Listen(addr string) (net.Listener, error) {
	return winio.ListenPipe(addr, &winio.PipeConfig{
		InputBufferSize:  128,
		OutputBufferSize: 128,
	})
}

func (_Pipe) Dial(ctx context.Context, addr string) (net.Conn, error) {
	var timeout *time.Duration
	deadline, ok := ctx.Deadline()
	if ok {
		timeoutValue := deadline.Sub(time.Now())
		timeout = &timeoutValue
	}
	return winio.DialPipe(addr, timeout)
}
