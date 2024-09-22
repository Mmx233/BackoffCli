package backoff

import (
	"context"
	"errors"
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
	err := instance.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = instance.Start(ctx)
	if !errors.Is(err, &ErrorAlreadyRunning{}) {
		t.Fatal("expected ErrorAlreadyRunning, got:", err)
	}

	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*5)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("instance should not be running")
		case <-time.After(time.Millisecond):
			if instance.Running.Load() {
				continue
			}
		}
		break
	}
}
