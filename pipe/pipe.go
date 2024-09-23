package pipe

import (
	"context"
	"net"
	"net/http"
)

type Pipe interface {
	Addr(name string) string
	Listen(addr string) (net.Listener, error)
	Dial(ctx context.Context, addr string) (net.Conn, error)
}

func HttpListen(listener net.Listener, server *http.Server) error {
	server.SetKeepAlivesEnabled(false)
	return server.Serve(listener)
}
