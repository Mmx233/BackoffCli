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
	Fn      Fn
	Running *atomic.Bool
}

type Fn func(ctx context.Context) error
type HealthChecker func(ctx context.Context) <-chan error

type Conf struct {
	Logger          *log.Logger
	DisableRecovery bool

	// HealthChecker func will be called while waiting for Fn returning errors.
	// Once Fn returned anything, the context passed to NewHealthChecker will be canceled.
	// If error chan return nil, wait time will be reset. Otherwise, the context passed
	// to Fn and HealthChecker will be canceled.
	HealthChecker HealthChecker

	// InitialDuration means initial wait time, default 1 second
	InitialDuration time.Duration
	// MaxDuration means maximum retry wait time, default 20 minutes
	MaxDuration time.Duration
	// MaxRetry, default unlimited
	MaxRetry uint

	// $Next = ($Last + InterConstFactor) * (2 ^ ExponentFactor) + OuterConstFactor

	// ExponentFactor default 1
	ExponentFactor   int
	InterConstFactor time.Duration
	OuterConstFactor time.Duration
}

// New backoff instance with default values
func New(fn func(ctx context.Context) error, c Conf) Backoff {
	if c.InitialDuration == 0 {
		c.InitialDuration = time.Second
	}
	if c.MaxDuration == 0 {
		c.MaxDuration = time.Minute * 20
	}
	if c.ExponentFactor <= 0 {
		c.ExponentFactor = 1
	}
	return NewInstance(fn, c)
}

// NewInstance creates instance without default configs.
// Running with all zero values will lead to unexpected behavior.
// Consider to use New or set values by hand.
func NewInstance(fn func(ctx context.Context) error, conf Conf) Backoff {
	if conf.Logger == nil {
		conf.Logger = log.New()
	}
	return Backoff{
		Config:  conf,
		Fn:      fn,
		Running: &atomic.Bool{},
	}
}

func (b Backoff) Start(ctx context.Context) error {
	if b.Running.CompareAndSwap(false, true) {
		go func() {
			defer b.Running.Store(false)
			_ = b.Run(ctx)
		}()
		return nil
	}
	return &ErrorAlreadyRunning{}
}

func (b Backoff) _CallHealthCheck(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	resetWait, cancelFn := CtxResetWait{}.Must(ctx), CtxCancelFn{}.Must(ctx)

	healthCheckChan := b.Config.HealthChecker(ctx)

	go func() {
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				return
			case err := <-healthCheckChan:
				if err != nil {
					b.Config.Logger.Warnf("health check failed: %v", err)
					cancelFn()
					return
				}
				// call reset wait
				select {
				case <-ctx.Done():
					return
				case resetWait <- struct{}{}:
				}
			}
		}
	}()
}

func (b Backoff) _CallFn(ctx context.Context) <-chan error {
	ctx, cancel := context.WithCancel(ctx)
	// set capacity to 1 to avoid goroutine leak
	errChan := make(chan error, 1)
	ctx = CtxCancelFn{}.Set(ctx, cancel)

	go func() {
		defer cancel()

		if !b.Config.DisableRecovery {
			defer func() {
				if p := recover(); p != nil {
					if len(errChan) == 0 {
						var buf [4096]byte
						errChan <- &ErrorPanic{
							Reason: p,
							Stack:  string(buf[:runtime.Stack(buf[:], false)]),
						}
					}
				}
			}()
		}

		// We need to wait for Fn returning an error anyway.
		// If context is canceled by HealthCheck or parent,
		// Fn should terminate waiting on itself.
		errChan <- b.Fn(ctx)
	}()

	if b.Config.HealthChecker != nil {
		b._CallHealthCheck(ctx)
	}
	return errChan
}

func (b Backoff) Run(ctx context.Context) error {
	logger := b.Config.Logger.WithContext(ctx)

	retry := b.Config.MaxRetry
	if retry != 0 {
		// add 1 to ignore first try
		retry += 1
	}

	wait := b.Config.InitialDuration

	for {
		var resetWait = make(chan struct{})
		ctx := CtxResetWait{}.Set(ctx, resetWait)

		errChan := b._CallFn(ctx)

	waitFn:
		var err error
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-resetWait:
			wait = b.Config.InitialDuration
			goto waitFn
		case err = <-errChan:
			if err == nil {
				return nil
			}
			// break select
		}

		{
			logger := logger
			if retry != 0 {
				logger = logger.WithFields(log.Fields{
					"rest": retry - 1,
				})
			}
			logger = logger.WithFields(log.Fields{
				"wait": fmt.Sprintf("%.0fs", wait.Seconds()),
			})
			logger.Errorf("failed with error: %v", err)

		}

		if retry != 0 {
			retry -= 1
			if retry == 0 {
				logger.Errorln("max retry exceeded")
				return &ErrorMaxRetryExceeded{LastError: err}
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
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
}
