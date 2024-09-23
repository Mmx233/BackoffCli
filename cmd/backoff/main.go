package main

import (
	"context"
	"github.com/Mmx233/BackoffCli/backoff"
	"github.com/Mmx233/BackoffCli/pipe"
)

func main() {
	backoff.New(func(ctx context.Context) error {
		return nil
	}, backoff.Conf{})
	pipe.New()
}
