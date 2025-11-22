//go:build darwin
// +build darwin

package cli

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/veil-net/veilnet"
)

type conflux struct {
	anchor           *veilnet.Anchor
	api              *API
	metricsServer    *http.Server

	anchorMutex sync.Mutex
	anchorOnce  sync.Once
}

func newConflux() *conflux {
	c := &conflux{}
	c.api = newAPI(c)
	return c
}

func (c *conflux) Run() error {

	return c.api.Run()
}

func (c *conflux) Install() error {
	return nil
}

func (c *conflux) Start() error {
	return nil
}

func (c *conflux) Stop() error {
	return nil
}

func (c *conflux) Remove() error {
	return nil
}

func (c *conflux) Status() (bool, error) {
	return false, nil
}

func (c *conflux) StartVeilNet(apiBaseURL, anchorToken string, portal bool) error {

	// Lock the anchor mutex
	c.anchorMutex.Lock()
	defer c.anchorMutex.Unlock()

	// initialize the anchor once
	c.anchorOnce = sync.Once{}

	//Close existing anchor if any (defensive cleanup)
	if c.anchor != nil {
		c.anchor.Stop()
		c.anchor = nil
	}

	// Create the anchor
	c.anchor = veilnet.NewAnchor()
	err := c.anchor.Start(apiBaseURL, anchorToken, false)
	if err != nil {
		return err
	}

	// Link the anchor to the TUN device
	err = c.anchor.LinkWithTUN("veilnet", 1500)
	if err != nil {
		return err
	}

	// Close existing metrics server
	if c.metricsServer != nil {
		c.metricsServer.Shutdown(context.Background())
		c.metricsServer = nil
	}

	// Start the metrics server
	c.metricsServer = &http.Server{
		Addr:    ":9090",
		Handler: c.anchor.Metrics.GetHandler(),
	}
	go func() {
		if err := c.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			veilnet.Logger.Sugar().Errorf("metrics server error: %v", err)
		}
	}()

	return nil
}

func (c *conflux) StopVeilNet() {

	c.anchorOnce.Do(func() {

		// Lock the anchor mutex
		c.anchorMutex.Lock()
		defer c.anchorMutex.Unlock()

		// Stop the metrics server
		if c.metricsServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := c.metricsServer.Shutdown(ctx); err != nil {
				veilnet.Logger.Sugar().Errorf("failed to stop metrics server: %v", err)
			}
			c.metricsServer = nil
		}

		// Stop the anchor
		if c.anchor != nil {
			c.anchor.Stop()
			c.anchor = nil
		}
	})
}

func (c *conflux) GetAnchor() *veilnet.Anchor {
	if c.anchor == nil {
		return nil
	}
	return c.anchor
}
