package cli

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/veil-net/conflux/anchor"
	"github.com/veil-net/conflux/service"
)

type Register struct {
	RegistrationToken string   `short:"t" help:"The registration token" env:"VEILNET_REGISTRATION_TOKEN" json:"registration_token"`
	Rift              bool     `short:"r" help:"Enable rift mode, default: false" default:"false" env:"VEILNET_RIFT" json:"rift"`
	Guardian          string   `help:"The Guardian URL (Authentication Server), default: https://guardian.veilnet.app" default:"https://guardian.veilnet.app" env:"VEILNET_GUARDIAN" json:"guardian"`
	Veil              string   `help:"The veil URL, default: nats.veilnet.app" default:"nats.veilnet.app" env:"VEILNET_VEIL" json:"veil"`
	VeilPort          int      `help:"The veil port, default: 30422" default:"30422" env:"VEILNET_VEIL_PORT" json:"veil_port"`
	Tag               string   `help:"The tag for the conflux" env:"VEILNET_CONFLUX_TAG" json:"tag"`
	Cidr              string   `help:"The CIDR of the conflux" env:"VEILNET_CONFLUX_CIDR" json:"cidr"`
	JWT               string   `help:"The JWT for the conflux" env:"VEILNET_CONFLUX_JWT" json:"jwt"`
	JWKS_url          string   `help:"The JWKS URL for the conflux" env:"VEILNET_CONFLUX_JWKS_URL" json:"jwks_url"`
	Audience          string   `help:"The audience for the conflux" env:"VEILNET_CONFLUX_AUDIENCE" json:"audience"`
	Issuer            string   `help:"The issuer for the conflux" env:"VEILNET_CONFLUX_ISSUER" json:"issuer"`
	Taints            []string `help:"Create taints for the conflux and return a token for each taint for other conflux to join" env:"VEILNET_CONFLUX_TAINTS" json:"taints"`
	JoinTaints       []string `help:"Conflux will use these taints to form a cluster with other conflux" env:"VEILNET_CONFLUX_JOIN_TAINTS" json:"join_taints"`
}

type ConfluxToken struct {
	ConfluxID string `json:"conflux_id"`
	Token     string `json:"token"`
}

func (cmd *Register) Run() error {

	// Parse the command
	registrationRequest := &anchor.ResgitrationRequest{
		RegistrationToken: cmd.RegistrationToken,
		Guardian:          cmd.Guardian,
		Tag:               cmd.Tag,
		Cidr:              cmd.Cidr,
		JWT:               cmd.JWT,
		JWKS_url:          cmd.JWKS_url,
		Audience:          cmd.Audience,
		Issuer:            cmd.Issuer,
	}

	// Register the conflux
	registrationResponse, err := anchor.RegisterConflux(registrationRequest)
	if err != nil {
		Logger.Sugar().Errorf("failed to register conflux: %v", err)
		return err
	}

	// Create taints
	taints := make([]string, 0)
	if len(cmd.Taints) > 0 {
		for _, taint := range cmd.Taints {
			signature := sha256.Sum256([]byte(taint + registrationResponse.ConfluxID))
			encodedSignature := hex.EncodeToString(signature[:])
			taints = append(taints, encodedSignature)
			Logger.Sugar().Infof("created taint: %s, use this signature %s with --joint-taints to form a cluster with other conflux", taint, encodedSignature, encodedSignature)
		}
	}
	if len(cmd.JoinTaints) > 0 {
		for _, taint := range cmd.JoinTaints {
			taints = append(taints, taint)
		}
	}

	// Save the configuration
	config := &anchor.ConfluxConfig{
		ConfluxID: registrationResponse.ConfluxID,
		Token:     registrationResponse.Token,
		Guardian:  cmd.Guardian,
		Veil:      cmd.Veil,
		VeilPort:  cmd.VeilPort,
		Rift:      cmd.Rift,
		Taints:    taints,
	}

	// Save the configuration
	err = anchor.SaveConfig(config)
	if err != nil {
		Logger.Sugar().Errorf("failed to save configuration: %v", err)
		return err
	}

	// Install the service
	conflux := service.NewService()
	if err := conflux.Status(); err == nil {
		Logger.Sugar().Infof("reinstalling veilnet conflux service...")
		conflux.Remove()
	} else {
		Logger.Sugar().Infof("installing veilnet conflux service...")
	}
	err = conflux.Install()
	if err != nil {
		Logger.Sugar().Errorf("failed to install service: %v", err)
		return err
	}

	return nil
}
