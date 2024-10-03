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

func New(ctx context.Context, logger log.FieldLogger, quit func()) (DoSingleton, *Singleton) {
	var needSingleton = config.Config.Singleton
	var single = NewInstance(config.Config.Name, logger)
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

func NewInstance(name string, logger log.FieldLogger) *Singleton {
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
		Logger:   logger,
	}
}

type Singleton struct {
	Name string
	Addr string

	Pipe       pipe.Pipe
	HttpClient *http.Client
	Shutdown   func()

	Logger log.FieldLogger
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
		s.Logger.Debugln("quit other instance failed:", err)
	} else {
		time.Sleep(time.Second)
	}

	listener, err := s.Pipe.Listen(s.Addr)
	if err != nil {
		return err
	}

	go func() {
		s.Logger.Debugf("listening %s", s.Addr)
		server := &http.Server{}
		s.Shutdown()
		s.Shutdown = sync.OnceFunc(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				s.Logger.Warnln("http server shutdown failed:", err)
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
			s.Logger.Errorln("listen on pipe failed:", err)
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
