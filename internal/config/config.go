package config

import (
	"github.com/Mmx233/BackoffCli/backoff"
	"time"
)

var Config _Config

type _Config struct {
	Name string
	Path string

	DurationInitial time.Duration
	DurationMax     time.Duration

	RetryMax int

	FactorExponent   int
	FactorConstInter time.Duration
	FactorConstOuter time.Duration
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
