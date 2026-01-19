package service

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/veil-net/conflux/anchor"
)


type ServiceImpl struct {
}

func NewServiceImpl() *ServiceImpl {
	return &ServiceImpl{}
}

func (s *ServiceImpl) Run() {

	// Load the configuration
	config, err := anchor.LoadConfig()
	if err != nil {
		Logger.Sugar().Errorf("failed to load configuration: %v", err)
		return
	}

	// Initialize the anchor plugin
	anchor, client, err := anchor.NewAnchor()
	if err != nil {
		Logger.Sugar().Errorf("failed to initialize anchor plugin: %v", err)
		return
	}
	defer client.Kill()

	// Initialize the anchor instance
	err = anchor.CreateAnchor()
	if err != nil {
		Logger.Sugar().Errorf("failed to create anchor instance: %v", err)
		return
	}

	// Start the anchor
	err = anchor.StartAnchor(config.Guardian, config.Veil, config.VeilPort, config.Token, config.Portal)
	if err != nil {
		Logger.Sugar().Errorf("failed to start anchor: %v", err)
		return
	}

	// Create the TUN device
	err = anchor.CreateTUN("veilnet", 1500)
	if err != nil {
		Logger.Sugar().Errorf("failed to create TUN device: %v", err)
		return
	}

	// Attach the anchor with the TUN device
	err = anchor.AttachWithTUN()
	if err != nil {
		Logger.Sugar().Errorf("failed to attach anchor with TUN device: %v", err)
		return
	}

	// Wait for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-interrupt

	// Stop the anchor
	anchor.StopAnchor()

	// Destroy the anchor
	anchor.DestroyAnchor()

	// Destroy the TUN device
	anchor.DestroyTUN()
}