package singleton

import (
	"context"
	"errors"
	"github.com/Mmx233/BackoffCli/pipe"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

func New(name string) *Singleton {
	_pipe := pipe.New()
	addr := _pipe.Addr(name)

	return &Singleton{
		Name: name,
		Addr: addr,
		Pipe: _pipe,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
					return _pipe.Dial(addr)
				},
			},
		},
	}
}

type Singleton struct {
	Name       string
	Addr       string
	Pipe       pipe.Pipe
	HttpClient *http.Client

	cancel context.CancelFunc
}

func (s *Singleton) RequestExit(ctx context.Context) error {
	req, err := http.NewRequest("GET", "unix://backoff/exit", nil)
	if err != nil {
		return err
	}
	req.WithContext(ctx)

	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if string(data) != "ok" {
		return errors.New("response not ok")
	}
	return nil
}

func (s *Singleton) Run() error {
	if s.cancel != nil {
		s.cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	if err := s.RequestExit(ctx); err != nil {
		log.Debugln("exit other instance failed:", err)
	} else {
		time.Sleep(time.Second)
	}

	listener, err := s.Pipe.Listen(s.Addr)
	if err != nil {
		return err
	}

	go func() {
		if err := pipe.HttpListen(listener, &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/exit":
					_, _ = w.Write([]byte("ok"))
					go func() {
						time.Sleep(time.Millisecond * 50)
						os.Exit(0)
					}()
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}),
		}); err != nil {
			log.Fatalln("listen on pipe failed:", err)
		}
	}()
	return nil
}

func (s *Singleton) Close() error {
	s.HttpClient.CloseIdleConnections()
	s.cancel()
	return nil
}
