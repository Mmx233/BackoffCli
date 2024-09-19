package backoff

import (
	"context"
	"time"
)

type ProbeHealthCheckFn func(ctx context.Context) error

type ProbeHealthCheckerConfig struct {
	CheckInterval    time.Duration
	InitialDelay     time.Duration
	SuccessThreshold int
	FailureThreshold int
}

func NewProbeHealthChecker(fn ProbeHealthCheckFn, conf ProbeHealthCheckerConfig) HealthChecker {
	return func(ctx context.Context) <-chan error {
		errChan := make(chan error, 1)
		var success, failure int
		go func() {
			select {
			case <-ctx.Done():
				return
			case <-time.After(conf.CheckInterval):
			}

			time.Sleep(conf.InitialDelay)
			ticker := time.NewTicker(conf.CheckInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				case <-ticker.C:
					err := fn(ctx)
					if err != nil {
						failure++
						if failure >= conf.FailureThreshold {
							errChan <- err
							return
						}
						success = 0
					} else {
						success++
						if success >= conf.SuccessThreshold {
							errChan <- nil
							return
						}
						failure = 0
					}
				}
			}
		}()
		return errChan
	}
}
