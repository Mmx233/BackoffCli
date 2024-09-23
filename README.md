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
  -h, --[no-]help              Show context-sensitive help (also try --help-long
                               and --help-man).
      --duration.initial=0s    initial wait seconds
      --duration.max=5m        max wait seconds
      --retry.max=0            max retry, 0 means unlimited
      --factor.exponent=1      exponent factor
      --factor.const.inter=0s  inter const factor
      --factor.const.outer=0s  outer const factor
      --name=NAME              pipe name for singleton, default generate by path
      --[no-]singleton         run with singleton parton with unique name

Args:
  <path>  program to run
```

### Use as go library

```shell
~# go get github.com/Mmx233/BackoffCli/backoff
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