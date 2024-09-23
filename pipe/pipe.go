package pipe

import (
	"bufio"
	"net"
	"net/http"
)

type Pipe interface {
	Addr(name string) string
	Listen(addr string) (net.Listener, error)
	Dial(addr string) (net.Conn, error)
}

func HttpRequest(conn net.Conn, req *http.Request) (*http.Response, error) {
	err := req.Write(conn)
	if err != nil {
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func HttpListen(listener net.Listener, server *http.Server) error {
	server.SetKeepAlivesEnabled(false)
	return server.Serve(listener)
}
