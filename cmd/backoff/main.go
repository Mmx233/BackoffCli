package main

import (
	"context"
	"github.com/Mmx233/BackoffCli/backoff"
	"github.com/Mmx233/BackoffCli/internal/config"
	"github.com/Mmx233/BackoffCli/internal/singleton"
	"github.com/alecthomas/kingpin/v2"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path"
	"strings"
)

func main() {
	kingpin.MustParse(config.NewCommands().Parse(os.Args[1:]))
	if config.Config.Name == "" {
		config.Config.Name = strings.Split(path.Base(os.Args[0]), ".")[0]
	}

	single := singleton.New(config.Config.Name)
	defer single.Close()
	if err := single.Run(); err != nil {
		log.Fatalln(err)
	}

	backoffConf := config.Config.NewBackoffConf()
	backoffInstance := backoff.NewInstance(func(ctx context.Context) error {
		parts := strings.Fields(config.Config.Path)
		cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}, backoffConf)
	if err := backoffInstance.Run(context.Background()); err != nil {
		log.Fatalln(err)
	}
}
