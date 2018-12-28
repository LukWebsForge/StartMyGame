package manager

import (
	"start-my-game/lib/cloud"
)

// Can be one of 'off', 'setup', 'setup_error', 'active'
func (manager *Manager) GetServerStatus() string {
	if manager.Startup != nil && manager.Startup.Error {
		return "startup_error"
	}

	if manager.ActiveServer == nil {
		return cloud.StatusOff
	}

	switch manager.ActiveServer.Status {
	case cloud.StatusActive:
		return cloud.StatusActive
	case cloud.StatusStartup:
		return cloud.StatusStartup
	default:
		return cloud.StatusOff
	}
}
