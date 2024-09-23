package config

import (
	"github.com/alecthomas/kingpin/v2"
)

func NewCommands() *kingpin.Application {
	app := kingpin.New("backoff", "A command-line tool designed to implement and manage customizable backoff strategies for retrying failed operations efficiently..")
	app.HelpFlag.Short('h')

	app.Flag("duration.initial", "initial wait seconds").Default("0s").DurationVar(&Config.DurationInitial)
	app.Flag("duration.max", "max wait seconds").Default("5m").DurationVar(&Config.DurationMax)

	app.Flag("retry.max", "max retry, 0 means unlimited").Default("0").IntVar(&Config.RetryMax)

	app.Flag("factor.exponent", "exponent factor").Default("1").IntVar(&Config.FactorExponent)
	app.Flag("factor.const.inter", "inter const factor").Default("0s").DurationVar(&Config.FactorConstInter)
	app.Flag("factor.const.outer", "outer const factor").Default("0s").DurationVar(&Config.FactorConstInter)

	app.Arg("path", "program to run").Required()

	return app
}
