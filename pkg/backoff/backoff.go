package backoff

import (
	"context"
	"github.com/Mmx233/BackoffCli/tools"
	"sync/atomic"
	"time"
)

// Backoff 错误重试 积分退避算法
type Backoff struct {
	retry time.Duration

	conf    Conf
	running *atomic.Bool
}

type Conf struct {
	Content func(ctx context.Context) error

	// MaxRetryDelay maximum retry wait time, default 20 minutes
	MaxRetryDelay time.Duration
	// RetryGrowthFactor determines how many times the retry time doubles, default 1
	RetryGrowthFactor int
}

func New(c Conf) Backoff {
	if c.Content == nil {
		panic("content function required")
	}
	if c.MaxRetryDelay == 0 {
		c.MaxRetryDelay = time.Minute * 20
	}
	if c.RetryGrowthFactor <= 0 {
		c.RetryGrowthFactor = 1
	}

	return Backoff{
		retry:   time.Second,
		conf:    c,
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

// Worker
// 请注意,此处使用的是普通接收器,当 worker 重新运行时 retry 会被重置
func (a Backoff) Worker(ctx context.Context) {
	for {
		errChan := make(chan error)
		go func() {
			defer tools.Recover()
			errChan <- a.conf.Content(ctx)
		}()
		if err := <-errChan; err == nil {
			break
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(a.retry):
			// continue retry
		}

		if a.retry < a.conf.MaxRetryDelay {
			a.retry = a.retry << a.conf.RetryGrowthFactor
			if a.retry > a.conf.MaxRetryDelay {
				a.retry = a.conf.MaxRetryDelay
			}
		}
	}

	a.running.Store(false)
}
