package backoff

import (
	"context"
	"github.com/Mmx233/BackoffCli/tools"
	log "github.com/sirupsen/logrus"
	"sync/atomic"
	"time"
)

// Backoff 错误重试 积分退避算法
type Backoff struct {
	Config  Conf
	running *atomic.Bool
}

type Conf struct {
	Fn     func(ctx context.Context) error
	Logger *log.Logger

	// InitialDuration initial wait time, default 1 second
	InitialDuration time.Duration
	// MaxDuration maximum retry wait time, default 20 minutes
	MaxDuration time.Duration

	// $Next = ($Last + InterConstFactor) * (2 ^ ExponentFactor) + OuterConstFactor

	// ExponentFactor default 1
	ExponentFactor   int
	InterConstFactor time.Duration
	OuterConstFactor time.Duration
}

// New backoff instance with default values
func New(c Conf) Backoff {
	if c.InitialDuration == 0 {
		c.InitialDuration = time.Second
	}
	if c.MaxDuration == 0 {
		c.MaxDuration = time.Minute * 20
	}
	if c.ExponentFactor <= 0 {
		c.ExponentFactor = 1
	}
	return NewInstance(c)
}

func NewInstance(c Conf) Backoff {
	if c.Fn == nil {
		panic("content function required")
	}
	if c.Logger == nil {
		c.Logger = log.New()
	}
	return Backoff{
		Config:  c,
		running: &atomic.Bool{},
	}
}

func (a Backoff) Start(ctx context.Context) error {
	if a.running.CompareAndSwap(false, true) {
		go a.Worker(ctx)
		return nil
	}
	return &ErrorAlreadyRunning{}
}

func (a Backoff) Worker(ctx context.Context) {
	wait := a.Config.InitialDuration
	for {
		errChan := make(chan error)
		go func() {
			defer tools.Recover()
			errChan <- a.Config.Fn(ctx)
		}()
		if err := <-errChan; err == nil {
			break
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
			// continue retry
		}

		if wait < a.Config.MaxDuration {
			wait = (wait+a.Config.InterConstFactor)<<a.Config.ExponentFactor + a.Config.OuterConstFactor
			if wait > a.Config.MaxDuration {
				wait = a.Config.MaxDuration
			}
		}
	}

	a.running.Store(false)
}
