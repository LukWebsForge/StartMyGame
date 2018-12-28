package cloud

import (
	"fmt"
	"start-my-game/lib/config"
	"strings"
)

const (
	StatusStartup   = "startup"
	StatusActive    = "active"
	StatusOff       = "off"
	StatusDestroyed = "destroyed"
)

type Cloud interface {
	GetProvider() string
	GetSSHKey(fingerprint string) (int, error)
	GetSnapshot(name string) (*Snapshot, error)
	GetServer(name string) (*Server, error)
	StartServer(server *Server) error
	StopServer(server *Server) error
	CreateServer(options CreateOptions) (*Server, error)
	DestroyServer(server *Server) error
}

type Server struct {
	Name     string
	Id       int
	Ip       string
	Status   string
	Provider string
}

type Snapshot struct {
	Name string
	Id   int
}

type CreateOptions struct {
	Name     string
	Snapshot *Snapshot
	SshKey   int
	Machine  string
	Region   string
}

func NewCloud(config *config.Config) (Cloud, error) {
	provider := strings.ToLower(config.Cloud.Provider)
	token := config.Cloud.Token

	var cloud Cloud

	switch provider {
	case strings.ToLower(hetznerProvider):
		cloud = newHCloud(token)
	case strings.ToLower(digitalOceanProvider):
		cloud = newDoCloud(token)
	}

	if cloud != nil {
		// No reference, because the cloud is an interface
		return cloud, nil
	}

	return nil, fmt.Errorf("could provider with name '%v' not found", provider)
}
