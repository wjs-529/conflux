package cli

import (
	"fmt"
	"io"
	"net/http"

	"github.com/veil-net/conflux/service"
)

type Unregister struct {
}

func (cmd *Unregister) Run() error {

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
			for {
				resp, err := http.Get("http://127.0.0.1:1993/health")
				if err == nil && resp.StatusCode == http.StatusOK {
					break
				}
				Logger.Sugar().Warnf("Waiting for VeilNet Conflux service to be ready...")
			}
		}
	}

	// Create the request to API
	req, err := http.NewRequest("DELETE", "http://127.0.0.1:1993/unregister", nil)
	if err != nil {
		Logger.Sugar().Errorf("Failed to create request to API: %v", err)
		return err
	}

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
		Logger.Sugar().Errorf("Failed to unregister conflux: %s: %s", resp.Status, string(body))
		return fmt.Errorf("failed to unregister conflux: %s: %s", resp.Status, string(body))
	}

	Logger.Sugar().Infof("VeilNet Conflux unregistered successfully")
	
	return nil
}
