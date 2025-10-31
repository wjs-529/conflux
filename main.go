package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/veil-net/conflux/cli"
	"github.com/veil-net/veilnet"
)

var version = "0.0.6"

func main() {
	// Parse the CLI arguments
	var cli cli.CLI
	ctx := kong.Parse(&cli, kong.Vars{"version": version})
	err := ctx.Run()
	if err != nil {
		veilnet.Logger.Sugar().Errorf("%v", err)
		os.Exit(1)
	}
}
