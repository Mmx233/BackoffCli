package backoff

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"runtime"
	"sync/atomic"
	"time"
)

type Backoff struct {
	Config  Conf
	running *atomic.Bool
}

type Conf struct {
	Logger *log.Logger

	Fn             func(ctx context.Context) error
	DisableRecover bool

	HealthCheck       func(ctx context.Context) <-chan error
	HealthCheckAlways bool

	// InitialDuration initial wait time, default 1 second
	InitialDuration time.Duration
	// MaxDuration maximum retry wait time, default 20 minutes
	MaxDuration time.Duration
	// MaxRetry default unlimited
	MaxRetry uint

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
		go func() {
			defer b.running.Store(false)
			_ = b.Run(ctx)
		}()
		return nil
	}
	return &ErrorAlreadyRunning{}
}

func (b Backoff) Run(ctx context.Context) error {
	logger := b.Config.Logger.WithContext(ctx)

	retry := b.Config.MaxRetry
	if retry != 0 {
		retry += 1
	}

	wait := b.Config.InitialDuration
	errChan := make(chan error)

	for {
		go func() {
			if !b.Config.DisableRecover {
				defer func() {
					if p := recover(); p != nil {
						var buf [4096]byte
						errChan <- &ErrorPanic{
							Reason: p,
							Stack:  string(buf[:runtime.Stack(buf[:], false)]),
						}
					}
				}()
			}
			errChan <- b.Config.Fn(ctx)
		}()

	healthCheck:
		var healthCheckChan = make(<-chan error)
		if b.Config.HealthCheck != nil && (b.Config.HealthCheckAlways || wait != b.Config.InitialDuration) {
			healthCheckChan = b.Config.HealthCheck(ctx)
		}

		var err error
		select {
		case err = <-healthCheckChan:
			if err == nil {
				wait = b.Config.InitialDuration
			} else {
				b.Config.Logger.Warnf("health check failed: %v", err)
			}
			goto healthCheck
		case err = <-errChan:
			if err == nil {
				return nil
			}
			// break select
		}

		{
			logger := logger.WithFields(log.Fields{
				"wait": fmt.Sprintf("%.0fs", wait.Seconds()),
			})
			if retry != 0 {
				logger = logger.WithField("retry", retry-1)
			}
			logger.Errorf("failed with error: %v", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
			// continue retry
		}

		if retry != 0 {
			retry -= 1
			if retry == 0 {
				logger.Errorln("max retry exceeded")
				return &ErrorMaxRetryExceeded{LastError: err}
			}
		}

		if wait < b.Config.MaxDuration {
			wait = (wait+b.Config.InterConstFactor)<<b.Config.ExponentFactor + b.Config.OuterConstFactor
			if wait > b.Config.MaxDuration {
				wait = b.Config.MaxDuration
			}
		}
	}
}
