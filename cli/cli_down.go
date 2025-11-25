package cli

import "os"

type Down struct{}

func (cmd *Down) Run() error {
	conflux := NewConflux()
	os.Setenv("VEILNET_GUARDIAN", "")
	os.Setenv("VEILNET_CONFLUX_TOKEN", "")
	os.Setenv("VEILNET_PORTAL", "false")
	conflux.Remove()
	return nil
}