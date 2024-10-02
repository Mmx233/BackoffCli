package singleton

import (
	"context"
	"errors"
	"github.com/Mmx233/BackoffCli/internal/config"
	"github.com/Mmx233/BackoffCli/pipe"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

type DoSingleton func() error

func New(ctx context.Context, quit func()) (DoSingleton, *Singleton) {
	var needSingleton = config.Config.Singleton
	var single = NewInstance(config.Config.Name)
	return func() error {
		if needSingleton {
			if err := single.Run(ctx, quit); err != nil {
				return err
			}
			needSingleton = false
		}
		return nil
	}, single
}

func NewInstance(name string) *Singleton {
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

func (s *Singleton) Run(ctx context.Context, quit func()) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if err := s.RequestExit(ctx); err != nil {
		log.Debugln("quit other instance failed:", err)
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
			quit()
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
		if err := pipe.HttpListen(listener, server); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorln("listen on pipe failed:", err)
			quit()
		}
	}()
	return nil
}

func (s *Singleton) Close() error {
	s.HttpClient.CloseIdleConnections()
	s.Shutdown()
	return nil
}
