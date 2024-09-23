package main

import (
	"github.com/Mmx233/BackoffCli/internal/config"
	_ "github.com/Mmx233/BackoffCli/pipe"
	"github.com/alecthomas/kingpin/v2"
	"os"
)

func main() {
	kingpin.MustParse(config.NewCommands().Parse(os.Args[1:]))
}
