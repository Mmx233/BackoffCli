package backoff

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBackoff_Start(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	instance := NewInstance(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
		}
		return nil
	}, Conf{})
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
