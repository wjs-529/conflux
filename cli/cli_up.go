package cli

type Up struct {
	Token    string `short:"t" help:"The conflux token, please keep it secret" env:"VEILNET_CONFLUX_TOKEN" json:"conflux_token"`
	Guardian string `help:"The Guardian URL (Authentication Server), default: https://guardian.veilnet.app" default:"https://guardian.veilnet.app" env:"VEILNET_GUARDIAN" json:"guardian"`
	Veil     string `help:"The veil URL, default: nats.veilnet.app" default:"nats.veilnet.app" env:"VEILNET_VEIL" json:"veil"`
	VeilPort int    `help:"The veil port, default: 30422" default:"30422" env:"VEILNET_VEIL_PORT" json:"veil_port"`
	Portal   bool   `short:"p" help:"Enable portal mode, default: false" default:"false" env:"VEILNET_PORTAL" json:"portal"`
}

func (cmd *Up) Run() error {
	return nil
}
