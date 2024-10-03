package main

import (
	"context"
	"errors"
	"github.com/Mmx233/BackoffCli/backoff"
	_backoff "github.com/Mmx233/BackoffCli/internal/backoff"
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

	_singleton, singletonInstance := singleton.New(ctx, quitProcess)
	defer singletonInstance.Shutdown()

	backoffConf, err := config.Config.NewBackoffConf(), error(nil)
	backoffConf.HealthChecker, err = _backoff.NewHealthCheckFn()
	if err != nil {
		log.Warnln("create health checker failed, proceed without health check:", err)
	}

	lastCmd := make(chan *exec.Cmd, 1)
	backoffInstance := backoff.NewInstance(_backoff.NewBackoffFn(lastCmd, _singleton), backoffConf)
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
