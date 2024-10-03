package config

import (
	"github.com/alecthomas/kingpin/v2"
)

func NewCommands() *kingpin.Application {
	app := kingpin.New("backoff", "A command-line tool designed to implement and manage customizable backoff strategies for retrying failed operations efficiently..")
	app.HelpFlag.Short('h')

	app.Flag("duration.initial", "initial wait time").Default("1s").DurationVar(&Config.DurationInitial)
	app.Flag("duration.max", "max wait time").Default("5m").DurationVar(&Config.DurationMax)

	app.Flag("retry.max", "max retry, 0 means unlimited").Default("0").IntVar(&Config.RetryMax)

	app.Flag("factor.exponent", "exponent factor").Default("1").IntVar(&Config.FactorExponent)
	app.Flag("factor.const.inter", "inter const factor").Default("0s").DurationVar(&Config.FactorConstInter)
	app.Flag("factor.const.outer", "outer const factor").Default("0s").DurationVar(&Config.FactorConstInter)

	app.Flag("probe.initial.delay", "probe health check initial delay").Default("1s").DurationVar(&Config.ProbeInitialDelay)
	app.Flag("probe.interval", "probe health check interval").Default("5s").DurationVar(&Config.ProbeInterval)
	app.Flag("probe.threshold.success", "probe health check success threshold").Default("1").IntVar(&Config.ProbeThresholdSuccess)
	app.Flag("probe.threshold.failure", "probe health check failure threshold").Default("5").IntVar(&Config.ProbeThresholdFailure)

	app.Flag("tcp.addr", "tcp health check addr").HintOptions("127.0.0.1:80").StringVar(&Config.TcpAddr)
	app.Flag("tcp.timeout", "tcp health check handshake timeout").Default("20s").DurationVar(&Config.TcpTimeout)

	app.Flag("http.url", "http health check url").HintOptions("https://example.com").StringVar(&Config.HttpUrl)
	app.Flag("http.method", "http health check request method").Default("GET").EnumVar(&Config.HttpMethod, "GET", "POST", "PUT", "DELETE", "PATCH")
	app.Flag("http.timeout", "http health check request timeout").Default("30s").DurationVar(&Config.HttpTimeout)
	app.Flag("http.insecure", "http health check skip ssl certificate verification").Default("false").BoolVar(&Config.HttpInsecure)
	app.Flag("http.headers", "http health check custom header").StringMapVar(&Config.HttpHeader)
	app.Flag("http.follow_redirect", "http health check follow redirect").Default("false").BoolVar(&Config.HttpFollowRedirect)
	app.Flag("http.status_code", "http health check target status code").IntVar(&Config.HttpStatusCode)
	app.Flag("http.keyword", "http health check target keyword").StringVar(&Config.HttpKeyword)

	app.Flag("name", "pipe name for singleton, default generate by path").StringVar(&Config.Name)
	app.Flag("singleton", "run with singleton parton with unique name").Default("false").BoolVar(&Config.Singleton)
	app.Arg("path", "program to run").Required().StringVar(&Config.Path)

	return app
}
