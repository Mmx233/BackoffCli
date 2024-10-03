package config

import (
	"github.com/Mmx233/BackoffCli/backoff"
	"time"
)

var Config _Config

type _Config struct {
	Name      string
	Path      string
	Singleton bool

	DurationInitial time.Duration
	DurationMax     time.Duration

	RetryMax int

	FactorExponent   int
	FactorConstInter time.Duration
	FactorConstOuter time.Duration

	ProbeInitialDelay     time.Duration
	ProbeInterval         time.Duration
	ProbeThresholdSuccess int
	ProbeThresholdFailure int

	TcpAddr    string
	TcpTimeout time.Duration

	HttpUrl            string
	HttpMethod         string
	HttpTimeout        time.Duration
	HttpHeader         map[string]string
	HttpInsecure       bool
	HttpFollowRedirect bool
	HttpStatusCode     int
	HttpKeyword        string
}

func (c _Config) NewBackoffConf() backoff.Conf {
	return backoff.Conf{
		InitialDuration:  c.DurationInitial,
		MaxDuration:      c.DurationMax,
		MaxRetry:         uint(c.RetryMax),
		ExponentFactor:   c.FactorExponent,
		InterConstFactor: c.FactorConstInter,
		OuterConstFactor: c.FactorConstOuter,
	}
}
