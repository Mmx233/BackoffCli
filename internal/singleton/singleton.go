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
	"sync"
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
				DialContext: func(ctx context.Context, _ string, _ string) (net.Conn, error) {
					return _pipe.Dial(ctx, addr)
				},
			},
		},
		Shutdown: func() {},
	}
}

type Singleton struct {
	Name string
	Addr string

	Pipe       pipe.Pipe
	HttpClient *http.Client
	Shutdown   func()
}

func (s *Singleton) RequestExit(ctx context.Context) error {
	req, err := http.NewRequest("GET", "http://backoff/exit", nil)
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
		return errors.New("http server response not correct")
	}
	return nil
}

func (s *Singleton) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

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
		log.Debugf("listening %s", s.Addr)
		server := &http.Server{}
		s.Shutdown()
		s.Shutdown = sync.OnceFunc(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				log.Warnln("http server shutdown failed:", err)
			}
			os.Exit(0)
		})
		server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/exit":
				_, _ = w.Write([]byte("ok"))
				go s.Shutdown()
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		})
		if err := pipe.HttpListen(listener, server); err != nil {
			log.Fatalln("listen on pipe failed:", err)
		}
	}()
	return nil
}

func (s *Singleton) Close() error {
	s.HttpClient.CloseIdleConnections()
	s.Shutdown()
	return nil
}
