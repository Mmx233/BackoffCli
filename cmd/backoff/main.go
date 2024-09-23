package main

import (
	"context"
	"github.com/Mmx233/BackoffCli/backoff"
)

func main() {
	backoff.New(func(ctx context.Context) error {
		return nil
	}, backoff.Conf{})
}
