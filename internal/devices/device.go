package devices

import "github.com/kurlmarx/romwrangler/internal/systems"

// ConnectionInfo holds connection details for a device.
type ConnectionInfo struct {
	Host     string
	Port     int
	User     string
	Password string
}

// Device is the interface for target gaming devices.
type Device interface {
	Name() string
	ROMBasePath() string
	FolderForSystem(id systems.SystemID) (string, bool)
	SupportedSystems() []systems.SystemID
	DefaultConnection() ConnectionInfo
}
