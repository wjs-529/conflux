package cli

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Unregister struct {
}

func (cmd *Unregister) Run() error {
	baseURL, err := url.Parse("http://localhost:1993")
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %v", err)
	}

	baseURL.Path = "/unregister"

	req, err := http.NewRequest("POST", baseURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
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
		return fmt.Errorf("failed to unregister conflux: %s: %s", resp.Status, string(body))
	}

	return nil
}
