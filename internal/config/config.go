package config

import "time"

var Config _Config

type _Config struct {
	DurationInitial time.Duration
	DurationMax     time.Duration

	RetryMax int

	FactorExponent   int
	FactorConstInter time.Duration
	FactorConstOuter time.Duration
}
