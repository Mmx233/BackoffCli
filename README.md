# Backoff CLI

[![Lisense](https://img.shields.io/github/license/Mmx233/BackoffCli)](https://github.com/Mmx233/BackoffCli/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/Mmx233/BackoffCli?color=blueviolet&include_prereleases)](https://github.com/Mmx233/BackoffCli/releases)
[![GoReport](https://goreportcard.com/badge/github.com/Mmx233/BackoffCli)](https://goreportcard.com/report/github.com/Mmx233/BackoffCli)

```shell
~# go run .\cmd\backoff\ -h                      
usage: backoff [<flags>] <path>

A command-line tool designed to implement and manage customizable backoff
strategies for retrying failed operations efficiently..


Flags:
  -h, --[no-]help               Show context-sensitive help (also try
                                --help-long and --help-man).
      --duration.initial=1s     initial wait time
      --duration.max=5m         max wait time
      --retry.max=0             max retry, 0 means unlimited
      --factor.exponent=1       exponent factor
      --factor.const.inter=0s   inter const factor
      --factor.const.outer=0s   outer const factor
      --probe.initial.delay=1s  probe health check initial delay
      --probe.interval=5s       probe health check interval
      --probe.threshold.success=1
                                probe health check success threshold
      --probe.threshold.failure=5
                                probe health check failure threshold
      --tcp.addr=TCP.ADDR       tcp health check addr
      --tcp.timeout=20s         tcp health check handshake timeout
      --http.url=HTTP.URL       http health check url
      --http.method=GET         http health check request method
      --http.timeout=30s        http health check request timeout
      --[no-]http.insecure      http health check skip ssl certificate
                                verification
      --http.headers=HTTP.HEADERS ...
                                http health check custom header
      --[no-]http.follow_redirect
                                http health check follow redirect
      --http.status_code=HTTP.STATUS_CODE
                                http health check target status code
      --http.keyword=HTTP.KEYWORD
                                http health check target keyword
      --name=NAME               pipe name for singleton, default generate by
                                path
      --[no-]singleton          run with singleton parton with unique name

Args:
  <path>  program to run

```

### Backoff Wait Time Calculating Logic

```
$Wait = duration.initial

for {
    if Fn_Succuess {
        quit
    }
    if HealthCheckExist && HealthCheckSuccess {
        $Wait = duration.initial
    } else {
        $Wait = ($Wait + factor.const.inter) * (2 ^ factor.exponent) + factor.const.outer
        $Wait = min($Wait, duration.max)
    }
    
    sleep($Wait)
}
```

### Use as go library

```shell
go get github.com/Mmx233/BackoffCli/backoff
```

```go
package main

import (
	"context"
	"github.com/Mmx233/BackoffCli/backoff"
	"time"
)

func main() {
	instance := backoff.New(func(ctx context.Context) error {
		// put logic here
		return nil
	}, backoff.Conf{
		Logger:           nil, // logrus logger
		DisableRecovery:  false,
		HealthChecker:    func(ctx context.Context) <-chan error {
			// health check logic
		},
		InitialDuration:  time.Second,
		MaxDuration:      time.Second*10,
		MaxRetry:         10,
		ExponentFactor:   1,
		InterConstFactor: time.Second,
		OuterConstFactor: time.Second,
	})

	if err := instance.Run(context.Background()); err != nil {
		panic(err)
	}
}

```