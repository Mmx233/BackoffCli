package backoff

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
	"time"
)

func TestBackoff_Recovery(t *testing.T) {
	t.Parallel()

	logger := log.New()
	logger.SetOutput(io.Discard)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	instance := NewInstance(func(ctx context.Context) error {
		panic("fn panic")
		return nil
	}, Conf{
		Logger:   logger,
		MaxRetry: 1,
	})

	var errorMaxRetry *ErrorMaxRetryExceeded
	require.ErrorAs(t, instance.Run(ctx), &errorMaxRetry)
	var errorPanic *ErrorPanic
	assert.ErrorAs(t, errorMaxRetry.LastError, &errorPanic)
}

func TestBackoff_HealthCheck(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.New()
	logger.SetOutput(io.Discard)

	ping := make(chan struct{})
	instance := NewInstance(func(ctx context.Context) error {
		ping <- struct{}{}
		select {
		case <-ctx.Done():
			cancel()
		case <-time.After(time.Second):
			t.Error("health check failure not causing context canceled")
			cancel()
		}
		return nil
	}, Conf{
		Logger: logger,
		HealthChecker: func(ctx context.Context) <-chan error {
			errChan := make(chan error, 1)
			go func() {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
					require.Equal(t, 1, len(ping), "instance fn not ran before healthcheck")
					return
				case <-ping:
				}
				errChan <- assert.AnError
			}()
			return errChan
		},
	})

	_ = instance.Run(ctx)
}

func TestBackoff_NextWait(t *testing.T) {
	instance := NewInstance(func(ctx context.Context) error {
		return nil
	}, Conf{
		InitialDuration:  time.Second,
		MaxDuration:      time.Minute * 5,
		ExponentFactor:   2,
		InterConstFactor: time.Second,
		OuterConstFactor: time.Second,
	})

	assert.Equal(t, time.Second*9, instance.NextWait(instance.Config.InitialDuration))
	assert.Equal(t, time.Second*5, instance.NextWait(0))

	assert.Equal(t, instance.Config.MaxDuration, instance.NextWait(time.Minute*2))
}
