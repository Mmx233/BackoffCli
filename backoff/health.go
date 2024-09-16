package backoff

import "time"

type HealthChecker struct {
	Config HealthCheckerConfig
}

type HealthCheckerConfig struct {
	CheckInterval    time.Duration
	InitialDelay     time.Duration
	SuccessThreshold int
	FailureThreshold int
}

func NewHealthChecker() {

}
