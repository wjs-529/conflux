package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/veil-net/conflux/service"
)

type Register struct {
	RegistrationToken string `short:"t" help:"The registration token" env:"VEILNET_REGISTRATION_TOKEN" json:"registration_token"`
	Guardian          string `help:"The Guardian URL (Authentication Server), default: https://guardian.veilnet.app" default:"https://guardian.veilnet.app" env:"VEILNET_GUARDIAN" json:"guardian"`
	Veil              string `help:"The veil URL, default: nats.veilnet.app" default:"nats.veilnet.app" env:"VEILNET_VEIL" json:"veil"`
	VeilPort          int    `help:"The veil port, default: 30422" default:"30422" env:"VEILNET_VEIL_PORT" json:"veil_port"`
	Portal            bool   `short:"p" help:"Enable portal mode, default: false" default:"false" env:"VEILNET_PORTAL" json:"portal"`
	Teams             string `help:"The teams to be forwarded by the conflux, separated by comma, e.g. team1,team2" env:"VEILNET_CONFLUX_TEAMS" json:"teams"`
	Tag               string `help:"The tag for the conflux" env:"VEILNET_CONFLUX_TAG" json:"tag"`
	Cidr              string `help:"The CIDR of the conflux" env:"VEILNET_CONFLUX_CIDR" json:"cidr"`
}

type ConfluxToken struct {
	ConfluxID string `json:"conflux_id"`
	Token     string `json:"token"`
}

func (cmd *Register) Run() error {

	// Check if the service is running
	conflux := service.NewService()
	err := conflux.Status()
	if err != nil {
		Logger.Sugar().Warnf("VeilNet Conflux service is not installed, installing...")
		err = conflux.Install()
		if err != nil {
			Logger.Sugar().Errorf("Failed to install VeilNet Conflux service: %v", err)
			return err
		} else {
			Logger.Sugar().Warnf("Waiting for VeilNet Conflux service to be ready...")
			for {
				resp, err := http.Get("http://127.0.0.1:1993/health")
				if err == nil && resp.StatusCode == http.StatusOK {
					break
				}
			}
		}
	}

	// Create the request to API
	body, err := json.Marshal(cmd)
	if err != nil {
		Logger.Sugar().Errorf("Failed to marshal register command: %v", err)
		return err
	}

	// Create the request
	req, err := http.NewRequest("POST", "http://127.0.0.1:1993/register", bytes.NewBuffer(body))
	if err != nil {
		Logger.Sugar().Errorf("Failed to create request to API: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		Logger.Sugar().Errorf("Failed to make request to API: %v", err)
		return err
	}
	defer resp.Body.Close()

	// if the response is not 200, return an error
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			Logger.Sugar().Errorf("Failed to read response body: %v", err)
			return err
		}
		Logger.Sugar().Errorf("Failed to register conflux: %s: %s", resp.Status, string(body))
		return fmt.Errorf("failed to register conflux: %s: %s", resp.Status, string(body))
	}

	Logger.Sugar().Infof("VeilNet Conflux registered successfully")

	return nil
}
