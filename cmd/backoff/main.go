package main

import (
	"context"
	"errors"
	"github.com/Mmx233/BackoffCli/backoff"
	_backoff "github.com/Mmx233/BackoffCli/internal/backoff"
	"github.com/Mmx233/BackoffCli/internal/config"
	"github.com/Mmx233/BackoffCli/internal/singleton"
	"github.com/alecthomas/kingpin/v2"
	nested "github.com/antonfisher/nested-logrus-formatter"
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
	logger := log.New()
	logger.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	quit := make(chan os.Signal)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	quitProcess := func() {
		select {
		case <-ctx.Done():
		case quit <- syscall.SIGTERM:
		}
	}

	_singleton, singletonInstance := singleton.New(ctx, logger.WithField(config.LogKeyComponent, "singleton"), quitProcess)
	defer singletonInstance.Shutdown()

	backoffConf, err := config.Config.NewBackoffConf(logger.WithField(config.LogKeyComponent, "backoff")), error(nil)
	backoffConf.HealthChecker, err = _backoff.NewHealthCheckFn(logger.WithField(config.LogKeyComponent, "health_checker"))
	if err != nil {
		logger.Warnln("create health checker failed, proceed without health check:", err)
	}

	lastCmd := make(chan *exec.Cmd, 1)
	backoffInstance := backoff.NewInstance(_backoff.NewBackoffFn(lastCmd, _singleton), backoffConf)
	go func() {
		if err := backoffInstance.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Errorln("backoff run failed:", err)
			quitProcess()
		}
	}()

	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-quit
	logger.Infoln("Shutdown...")
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
