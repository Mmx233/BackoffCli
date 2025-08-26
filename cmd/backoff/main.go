package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/Mmx233/BackoffCli/backoff"
	_backoff "github.com/Mmx233/BackoffCli/internal/backoff"
	"github.com/Mmx233/BackoffCli/internal/config"
	"github.com/Mmx233/BackoffCli/internal/singleton"
	"github.com/alecthomas/kingpin/v2"
	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

func init() {
	kingpin.MustParse(config.NewCommands().Parse(os.Args[1:]))
	if config.Config.Name == "" && len(config.Config.Commands) != 0 {
		config.Config.Name = "backoff-" + strings.Split(path.Base(strings.ReplaceAll(config.Config.Commands[0], "\\", "/")), ".")[0]
	}
}

func main() {
	logger := log.New()
	logger.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)

	singletonInstance := singleton.New(ctx, logger.WithField(config.LogKeyComponent, "singleton"), cancel)
	defer singletonInstance.Close()

	backoffConf, err := config.Config.NewBackoffConf(logger.WithField(config.LogKeyComponent, "backoff")), error(nil)
	backoffConf.HealthChecker, err = _backoff.NewHealthCheckFn(logger.WithField(config.LogKeyComponent, "health_checker"))
	if err != nil {
		logger.Warnln("create health checker failed, proceed without health check:", err)
	}

	lastCmd := make(chan *exec.Cmd, 1)
	if len(config.Config.Commands) != 0 {
		backoffInstance := backoff.NewInstance(_backoff.NewBackoffFn(lastCmd), backoffConf)
		go func() {
			if err := backoffInstance.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				logger.Errorln("backoff run failed:", err)
			}
			cancel()
		}()
	} else {
		logger.Infoln("no commands, doing nothing")
		cancel()
	}

	select {
	case <-ctx.Done():
	case <-quit:
	}

	logger.Infoln("shutdown...")
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
