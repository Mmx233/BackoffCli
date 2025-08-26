package backoff

import (
	"context"
	"os"
	"os/exec"

	"github.com/Mmx233/BackoffCli/backoff"
	"github.com/Mmx233/BackoffCli/internal/config"
)

func NewBackoffFn(lastCmd chan *exec.Cmd) backoff.Fn {
	return func(ctx context.Context) error {
		select {
		case <-lastCmd:
		default:
		}

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		cmd := exec.CommandContext(ctx, config.Config.Commands[0], config.Config.Commands[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		lastCmd <- cmd
		return cmd.Run()
	}
}
