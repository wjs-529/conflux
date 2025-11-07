package cli

import "github.com/veil-net/veilnet"

type Conflux interface {

	// Run runs the conflux service
	Run() error

	// Install installs the conflux service
	Install() error

	// Start starts the conflux service
	Start() error

	// Stop stops the conflux service
	Stop() error

	// Remove removes the conflux service
	Remove() error

	// Status returns the status of the conflux service
	Status() (bool, error)

	// Metrics returns the metrics of the conflux service
	Metrics(name string) int

	// StartVeilNet starts the veilnet service
	StartVeilNet(apiBaseURL, anchorToken string, portal bool) error

	// StopVeilNet stops the veilnet service
	StopVeilNet()

	// GetAnchor returns the anchor of the conflux service
	GetAnchor() *veilnet.Anchor

	// CreateTUN creates a TUN device
	CreateTUN() error

	// CloseTUN closes the TUN device
	CloseTUN()

	// DetectHostGateway detects the host default gateway and interface
	DetectHostGateway() error

	// AddBypassRoutes adds bypass routes
	AddBypassRoutes()

	// RemoveBypassRoutes removes bypass routes
	RemoveBypassRoutes()
}

func NewConflux() Conflux {
	return newConflux()
}
