package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Up struct {
	Token    string `short:"t" help:"The conflux token, please keep it secret" env:"VEILNET_CONFLUX_TOKEN" json:"conflux_token"`
	Portal   bool   `short:"p" help:"Enable portal mode, default: false" default:"false" env:"VEILNET_PORTAL" json:"portal"`
	Guardian string `short:"g" help:"The Guardian URL (Authentication Server), default: https://guardian.veilnet.app" default:"https://guardian.veilnet.app" env:"VEILNET_GUARDIAN" json:"guardian"`
}

func (cmd *Up) Run() error {

	return up(cmd.Guardian, cmd.Token, cmd.Portal)

}

func up(guardian, token string, portal bool) error {
	baseURL, err := url.Parse("http://localhost:1993")
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %v", err)
	}

	baseURL.Path = "/up"

	request := Up{
		Token:    token,
		Portal:   portal,
		Guardian: guardian,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", baseURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}
		return fmt.Errorf("failed to start conflux: %s: %s", resp.Status, string(body))
	}

	return nil
}
