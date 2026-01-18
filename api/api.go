package api

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	hcplugin "github.com/hashicorp/go-plugin"
	"github.com/veil-net/conflux/anchor"
)

type API struct {
	config          *Config
	plugin          *hcplugin.Client
	anchorInterface anchor.Anchor

	mu   sync.Mutex
	once sync.Once
}

func NewAPI() *API {
	return &API{}
}

func (a *API) initializeAnchor() error {
	// Stop the plugin if it is running
	if a.plugin != nil {
		a.plugin.Kill()
		a.anchorInterface = nil
		a.plugin = nil
	}
	// Create the plugin
	anchor, client, err := anchor.NewAnchor()
	if err != nil {
		return err
	}
	a.anchorInterface = anchor
	a.plugin = client
	return nil
}

func (a *API) resetAnchor() error {
	if a.plugin != nil {
		a.plugin.Kill()
		a.anchorInterface = nil
		a.plugin = nil
	}
	return nil
}

func (a *API) Run() {
	// Create the anchor interface
	err := a.initializeAnchor()
	if err != nil {
		Logger.Sugar().Fatalf("failed to create anchor interface: %v", err)
	}

	// Load existing configuration
	existingConfig, err := LoadConfig()
	if err == nil {
		Logger.Sugar().Infof("loaded existing configuration: Conflux Tag: %s, Portal: %t", existingConfig.Tag, existingConfig.Portal)
		// Register the conflux
		registrationResponse, err := RegisterConflux(existingConfig)
		if err == nil {
			// Create the anchor
			err = a.anchorInterface.CreateAnchor()
			if err != nil {
				Logger.Sugar().Fatalf("failed to create anchor: %v", err)
			}
			// Start the anchor
			err = a.anchorInterface.StartAnchor(existingConfig.Guardian, existingConfig.Veil, existingConfig.VeilPort, registrationResponse.Token, existingConfig.Portal)
			if err == nil {
				// Create the TUN device
				err = a.anchorInterface.CreateTUN("veilnet", 1500)
				if err == nil {
					err = a.anchorInterface.AttachWithTUN()
					if err == nil {
						a.config = existingConfig
						err = SaveConfig(a.config)
						if err != nil {
							Logger.Sugar().Errorf("failed to save configuration: %v", err)
						}
					} else {
						a.resetAnchor()
						Logger.Sugar().Warnf("failed to link anchor with TUN device: %v", err)
					}
				} else {
					a.resetAnchor()
					Logger.Sugar().Warnf("failed to create TUN device: %v", err)
				}
			} else {
				a.resetAnchor()
				Logger.Sugar().Warnf("failed to start anchor: %v", err)
			}
		} else {
			Logger.Sugar().Warnf("failed to register conflux instance: %v", err)
		}
	} else {
		Logger.Sugar().Warnf("failed to load configuration: %v", err)
	}
}

func (a *API) Stop() {
	a.once.Do(func() {
		a.mu.Lock()
		defer a.mu.Unlock()
		if a.plugin != nil {
			a.plugin.Kill()
			a.anchorInterface = nil
			a.plugin = nil
		}
	})
}

func (a *API) handleRegister(c *fiber.Ctx) error {
	a.mu.Lock()
	defer a.mu.Unlock()

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
	registrationResponse, err := RegisterConflux(config)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to register conflux: %v", err),
		})
	}
	// Initialize the plugin
	err = a.initializeAnchor()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to initialize plugin: %v", err),
		})
	}
	// Create the anchor
	err = a.anchorInterface.CreateAnchor()
	if err != nil {
		a.resetAnchor()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to create anchor: %v", err),
		})
	}
	// Start the anchor
	err = a.anchorInterface.StartAnchor(config.Guardian, config.Veil, config.VeilPort, registrationResponse.Token, config.Portal)
	if err != nil {
		a.resetAnchor()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to start anchor: %v", err),
		})
	}
	// Create the TUN device
	err = a.anchorInterface.CreateTUN("veilnet", 1500)
	if err != nil {
		a.resetAnchor()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to create TUN device: %v", err),
		})
	}
	// Attach the anchor to the TUN device
	err = a.anchorInterface.AttachWithTUN()
	if err != nil {
		a.resetAnchor()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to attach anchor to TUN device: %v", err),
		})
	}
	a.config = config
	// Save the configuration
	err = SaveConfig(config)
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
	err := DeleteConfig()
	if err != nil {
		Logger.Sugar().Warnf("failed to delete configuration file: %v", err)
	}

	if a.plugin != nil {
		// Get the conflux ID
		confluxID, err := a.anchorInterface.GetID()
		if err != nil {
			Logger.Sugar().Errorf("failed to get conflux ID: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"detail": fmt.Sprintf("failed to get conflux ID: %v", err),
			})
		}
		// Unregister the conflux
		err = UnregisterConflux(a.config, confluxID)
		if err != nil {
			Logger.Sugar().Errorf("failed to unregister conflux: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"detail": fmt.Sprintf("failed to unregister conflux: %v", err),
			})
		}
		// Stop the anchor plugin
		a.resetAnchor()
	}
	// Clear the config
	a.config = nil
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
	// Initialize the plugin
	err = a.initializeAnchor()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to initialize plugin: %v", err),
		})
	}
	// Create the anchor
	err = a.anchorInterface.CreateAnchor()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to create anchor: %v", err),
		})
	}
	// Start the anchor
	err = a.anchorInterface.StartAnchor(config.Guardian, config.Veil, config.VeilPort, config.Token, config.Portal)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to start anchor: %v", err),
		})
	}
	// Create the TUN device
	err = a.anchorInterface.CreateTUN("veilnet", 1500)
	if err != nil {
		a.resetAnchor()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to create TUN device: %v", err),
		})
	}
	// Attach the anchor to the TUN device
	err = a.anchorInterface.AttachWithTUN()
	if err != nil {
		a.resetAnchor()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to link anchor with TUN device: %v", err),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func (a *API) handleDown(c *fiber.Ctx) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Reset the plugin
	err := a.resetAnchor()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to reset plugin: %v", err),
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func (a *API) handleHealth(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}
