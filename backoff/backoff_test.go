package backoff

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
	"time"
)

func TestBackoff_Start(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	instance := New(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
		}
		return nil
	}, Conf{})
	instance.Config.Logger.SetOutput(io.Discard)

	require.Nil(t, instance.Start(ctx), "start instance failed")
	require.ErrorIs(t, instance.Start(ctx), &ErrorAlreadyRunning{})

	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*5)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("instance should not still running")
		case <-time.After(time.Millisecond):
			if instance.Running.Load() {
				continue
			}
		}
		break
	}
}

func TestBackoff_Recovery(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	instance := NewInstance(func(ctx context.Context) error {
		panic("fn panic")
		return nil
	}, Conf{
		MaxRetry: 1,
	})
	instance.Config.Logger.SetOutput(io.Discard)

	var errorMaxRetry *ErrorMaxRetryExceeded
	require.ErrorAs(t, instance.Run(ctx), &errorMaxRetry)
	var errorPanic *ErrorPanic
	assert.ErrorAs(t, errorMaxRetry.LastError, &errorPanic)
}

func TestBackoff_HealthCheck(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ping := make(chan struct{})
	instance := NewInstance(func(ctx context.Context) error {
		ping <- struct{}{}
		select {
		case <-ctx.Done():
			cancel()
		case <-time.After(time.Second):
			require.FailNow(t, "health check failure not causing context canceled")
			cancel()
		}
		return nil
	}, Conf{
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
	instance.Config.Logger.SetOutput(io.Discard)

	_ = instance.Run(ctx)
}
