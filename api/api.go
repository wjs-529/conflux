package api

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/veil-net/veilnet"
	"github.com/veil-net/veilnet/tun"
)

type API struct {
	anchor *veilnet.Anchor
	tun    tun.Tun
	config *Config
	app    *fiber.App

	mu   sync.Mutex
	once sync.Once
}

func NewAPI() *API {
	app := fiber.New()
	return &API{
		app: app,
	}
}

func (a *API) Run() {

	// Load existing configuration
	existingConfig, err := loadConfig()
	if err == nil {
		Logger.Sugar().Infof("loaded existing configuration: Conflux Tag: %s, Portal: %t", existingConfig.Tag, existingConfig.Portal)
		// Register the conflux
		registrationResponse, err := registerConflux(existingConfig)
		if err == nil {
			// Start the anchor
			anchor := veilnet.NewAnchor()
			err = anchor.Start(existingConfig.Guardian, existingConfig.Veil, existingConfig.VeilPort, registrationResponse.Token, existingConfig.Portal)
			if err == nil {
				// Create the TUN device
				tun, err := tun.CreateTun("veilnet", 1500)
				if err == nil {
					a.tun = tun
					err = anchor.LinkWithTUN(a.tun)
					if err == nil {
						a.anchor = anchor
						a.config = existingConfig
					} else {
						anchor.Stop()
						Logger.Sugar().Warnf("failed to link anchor with TUN device: %v", err)
					}
				} else {
					anchor.Stop()
					Logger.Sugar().Warnf("failed to create TUN device: %v", err)
				}
			} else {
				anchor.Stop()
				Logger.Sugar().Warnf("failed to start anchor: %v", err)
			}
		} else {
			Logger.Sugar().Warnf("failed to register conflux instance: %v", err)
		}
	} else {
		Logger.Sugar().Warnf("failed to load configuration: %v", err)
	}

	// Register the API routes
	a.app.Post("/register", a.handleRegister)
	a.app.Delete("/unregister", a.handleUnregister)
	a.app.Post("/up", a.handleUP)
	a.app.Delete("/down", a.handleDown)
	a.app.Get("/health", a.handleHealth)

	// Start the fiber app
	go func() {
		if err := a.app.Listen("127.0.0.1:1993"); err != nil {
			Logger.Sugar().Fatalf("VeilNet Conflux service has stopped: %v", err)
		}
	}()

	Logger.Sugar().Info("VeilNet Conflux service has started on port 1993")
}

func (a *API) Stop() {
	a.once.Do(func() {
		a.mu.Lock()
		defer a.mu.Unlock()
		if a.anchor != nil {
			a.anchor.Stop()
		}
		if a.app != nil {
			a.app.Shutdown()
		}
	})
}

func (a *API) handleRegister(c *fiber.Ctx) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if the service is already registered
	if a.config != nil && a.anchor != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"detail": "conflux already registered",
		})
	}

	// Parse the request body
	body := c.Body()
	config := &Config{}
	err := json.Unmarshal(body, config)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"detail": "failed to parse request body",
		})
	}
	// Register the conflux
	registrationResponse, err := registerConflux(config)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to register conflux: %v", err),
		})
	}
	// Stop the existing anchor
	if a.anchor != nil {
		a.anchor.Stop()
	}
	// Close the existing TUN device
	if a.tun != nil {
		a.tun.Close()
	}
	// Start the anchor
	anchor := veilnet.NewAnchor()
	err = anchor.Start(config.Guardian, config.Veil, config.VeilPort, registrationResponse.Token, config.Portal)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to start anchor: %v", err),
		})
	}
	// Create the TUN device
	tun, err := tun.CreateTun("veilnet", 1500)
	if err == nil {
		a.tun = tun
	} else {
		anchor.Stop()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to create TUN device: %v", err),
		})
	}
	// Create the TUN device
	err = anchor.LinkWithTUN(a.tun)
	if err != nil {
		anchor.Stop()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to create TUN device: %v", err),
		})
	}
	// Set current anchor and config
	a.anchor = anchor
	a.config = config
	// Save the configuration
	err = saveConfig(config)
	if err != nil {
		Logger.Sugar().Warnf("failed to save configuration: %v", err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (a *API) handleUnregister(c *fiber.Ctx) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.config == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"detail": "no configuration found, this instance is not registered",
		})
	}

	// Delete the configuration file
	err := deleteConfig()
	if err != nil {
		Logger.Sugar().Warnf("failed to delete configuration file: %v", err)
	}

	if a.anchor != nil {
		// Get the conflux ID
		confluxID, err := a.anchor.GetID()
		if err != nil {
			Logger.Sugar().Errorf("failed to get conflux ID: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"detail": fmt.Sprintf("failed to get conflux ID: %v", err),
			})
		}
		// Unregister the conflux
		err = unregisterConflux(a.config, confluxID)
		if err != nil {
			Logger.Sugar().Errorf("failed to unregister conflux: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"detail": fmt.Sprintf("failed to unregister conflux: %v", err),
			})
		}
		// Stop the anchor
		a.anchor.Stop()
		// Close the TUN device
		if a.tun != nil {
			a.tun.Close()
		}
	}
	// Clear the config
	a.config = nil
	a.anchor = nil
	return c.SendStatus(fiber.StatusOK)
}

func (a *API) handleUP(c *fiber.Ctx) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Parse the request body
	body := c.Body()
	config := &ConfluxConfig{}
	err := json.Unmarshal(body, config)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"detail": "failed to parse request body",
		})
	}
	// Stop the existing anchor
	if a.anchor != nil {
		a.anchor.Stop()
	}
	// Close the existing TUN device
	if a.tun != nil {
		a.tun.Close()
	}
	// Start the anchor
	anchor := veilnet.NewAnchor()
	err = anchor.Start(config.Guardian, config.Veil, config.VeilPort, config.Token, config.Portal)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to start anchor: %v", err),
		})
	}
	// Create the TUN device
	tun, err := tun.CreateTun("veilnet", 1500)
	if err == nil {
		a.tun = tun
	} else {
		anchor.Stop()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to create TUN device: %v", err),
		})
	}
	// Create the TUN device
	err = anchor.LinkWithTUN(a.tun)
	if err != nil {
		anchor.Stop()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to create TUN device: %v", err),
		})
	}
	// Set current anchor and config
	a.anchor = anchor
	return c.SendStatus(fiber.StatusOK)
}

func (a *API) handleDown(c *fiber.Ctx) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Stop the anchor
	if a.anchor != nil {
		a.anchor.Stop()
	}
	// Close the TUN device
	if a.tun != nil {
		a.tun.Close()
	}
	// Clear the anchor
	a.anchor = nil
	return c.SendStatus(fiber.StatusOK)
}

func (a *API) handleHealth(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}
