package backoff

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"runtime"
	"sync/atomic"
	"time"
)

// Backoff 错误重试 积分退避算法
type Backoff struct {
	Config  Conf
	running *atomic.Bool
}

type Conf struct {
	Fn             func(ctx context.Context) error
	Logger         *log.Logger
	DisableRecover bool

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

func NewInstance(conf Conf) Backoff {
	if conf.Fn == nil {
		panic("content function required")
	}
	if conf.Logger == nil {
		conf.Logger = log.New()
	}
	return Backoff{
		Config:  conf,
		running: &atomic.Bool{},
	}
}

func (b Backoff) Start(ctx context.Context) error {
	if b.running.CompareAndSwap(false, true) {
		go b.Worker(ctx)
		return nil
	}
	return &ErrorAlreadyRunning{}
}

func (b Backoff) Worker(ctx context.Context) {
	defer b.running.Store(false)
	logger := b.Config.Logger.WithContext(ctx)
	logger.Debugln("worker start")

	wait := b.Config.InitialDuration
	for {
		errChan := make(chan error)
		go func() {
			if !b.Config.DisableRecover {
				defer func() {
					if p := recover(); p != nil {
						var buf [4096]byte
						errChan <- &ErrorPanic{
							Err:   p,
							Stack: string(buf[:runtime.Stack(buf[:], false)]),
						}
					}
				}()
			}
			errChan <- b.Config.Fn(ctx)
		}()
		if err := <-errChan; err == nil {
			break
		} else {
			logger.WithFields(log.Fields{
				"wait": fmt.Sprintf("%.0fs", wait.Seconds()),
			}).Errorf("failed with error: %v", err)
		}

		select {
		case <-ctx.Done():
			logger.Infoln("task canceled")
			return
		case <-time.After(wait):
			// continue retry
		}

		if wait < b.Config.MaxDuration {
			wait = (wait+b.Config.InterConstFactor)<<b.Config.ExponentFactor + b.Config.OuterConstFactor
			if wait > b.Config.MaxDuration {
				wait = b.Config.MaxDuration
			}
		}
	}

	logger.Debugln("worker quit")
}
