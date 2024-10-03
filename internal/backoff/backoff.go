package backoff

import (
	"context"
	"github.com/Mmx233/BackoffCli/backoff"
	"github.com/Mmx233/BackoffCli/internal/config"
	"github.com/Mmx233/BackoffCli/internal/singleton"
	"os"
	"os/exec"
	"strings"
)

func NewBackoffFn(lastCmd chan *exec.Cmd, _singleton singleton.DoSingleton) backoff.Fn {
	return func(ctx context.Context) error {
		if err := _singleton(); err != nil {
			return err
		}

		select {
		case <-lastCmd:
		default:
		}

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		parts := strings.Fields(config.Config.Path)
		cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		lastCmd <- cmd
		return cmd.Run()
	}
}
