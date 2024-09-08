package backoff

import (
	"context"
	"github.com/Mmx233/BackoffCli/tools"
	"sync/atomic"
	"time"
)

// Backoff 错误重试 积分退避算法
type Backoff struct {
	Content func(ctx context.Context) error

	retry    time.Duration
	maxRetry time.Duration

	running *atomic.Bool
}

type Conf struct {
	Content func(ctx context.Context) error
	// 最大重试等待时间
	MaxRetryDelay time.Duration
}

func New(c Conf) Backoff {
	if c.Content == nil {
		panic("content function required")
	}
	if c.MaxRetryDelay == 0 {
		c.MaxRetryDelay = time.Minute * 20
	}

	return Backoff{
		Content:  c.Content,
		retry:    time.Second,
		maxRetry: c.MaxRetryDelay,
		running:  &atomic.Bool{},
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
			errChan <- a.Content(ctx)
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

		if a.retry < a.maxRetry {
			a.retry = a.retry << 1
			if a.retry > a.maxRetry {
				a.retry = a.maxRetry
			}
		}
	}

	a.running.Store(false)
}
