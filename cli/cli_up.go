package cli

import (
	"os"
	"strconv"
)

type Up struct {
	Guardian string `short:"g" help:"The Guardian URL (Authentication Server), default: https://guardian.veilnet.app" default:"https://guardian.veilnet.app" env:"VEILNET_GUARDIAN" json:"guardian"`
	Token    string `short:"t" help:"The conflux token, please keep it secret" env:"VEILNET_CONFLUX_TOKEN" json:"conflux_token"`
	Portal   bool   `short:"p" help:"Enable portal mode, default: false" default:"false" env:"VEILNET_PORTAL" json:"portal"`
}

func (cmd *Up) Run() error {
	conflux := NewConflux()
	conflux.Remove()
	os.Setenv("VEILNET_GUARDIAN", cmd.Guardian)
	os.Setenv("VEILNET_CONFLUX_TOKEN", cmd.Token)
	os.Setenv("VEILNET_PORTAL", strconv.FormatBool(cmd.Portal))
	err := conflux.Install()
	if err != nil {
		return err
	}
	return nil
}
