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
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	kingpin.MustParse(config.NewCommands().Parse(os.Args[1:]))
	quit := make(chan os.Signal)

	var needSingleton atomic.Bool
	var single *singleton.Singleton
	if config.Config.Singleton {
		needSingleton.Store(true)
		single = singleton.New(config.Config.Name)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	backoffConf := config.Config.NewBackoffConf()
	backoffInstance := backoff.NewInstance(func(ctx context.Context) error {
		if needSingleton.Load() {
			if config.Config.Name == "" {
				config.Config.Name = "backoff-" + strings.Split(path.Base(strings.ReplaceAll(config.Config.Path, "\\", "/")), ".")[0]
			}
			if err := single.Run(ctx, func() {
				select {
				case <-ctx.Done():
				case quit <- syscall.SIGTERM:
				}
			}); err != nil {
				return err
			}
			needSingleton.Store(false)
		}

		parts := strings.Fields(config.Config.Path)
		cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}, backoffConf)

	go func() {
		if err := backoffInstance.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Fatalln("backoff run failed:", err)
		}
	}()

	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-quit
	log.Infoln("Shutdown...")
	cancel()
	if single != nil {
		single.Shutdown()
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			log.Warnln("wait goroutine quit timeout")
			return
		case <-time.After(time.Millisecond * 10):
			if runtime.NumGoroutine() <= 2 {
				return
			}
		}
	}
}
