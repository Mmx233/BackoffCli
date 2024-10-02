package main

import (
	"context"
	"errors"
	"github.com/Mmx233/BackoffCli/backoff"
	"github.com/Mmx233/BackoffCli/internal/config"
	"github.com/Mmx233/BackoffCli/internal/singleton"
	"github.com/alecthomas/kingpin/v2"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"
)

func init() {
	kingpin.MustParse(config.NewCommands().Parse(os.Args[1:]))
	if config.Config.Name == "" {
		config.Config.Name = "backoff-" + strings.Split(path.Base(strings.ReplaceAll(config.Config.Path, "\\", "/")), ".")[0]
	}
}

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

func main() {
	quit := make(chan os.Signal)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	quitProcess := func() {
		select {
		case <-ctx.Done():
		case quit <- syscall.SIGTERM:
		}
	}

	_singleton, single := singleton.New(ctx, quitProcess)
	defer single.Shutdown()
	lastCmd := make(chan *exec.Cmd, 1)
	backoffInstance := backoff.NewInstance(NewBackoffFn(lastCmd, _singleton), config.Config.NewBackoffConf())
	go func() {
		if err := backoffInstance.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Errorln("backoff run failed:", err)
			quitProcess()
		}
	}()

	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-quit
	log.Infoln("Shutdown...")
	cancel()

	select {
	case cmd := <-lastCmd:
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_, _ = cmd.Process.Wait()
		}
	default:
	}
}
