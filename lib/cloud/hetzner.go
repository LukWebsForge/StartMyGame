package cloud

import (
	"context"
	"fmt"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"strings"
)

// https://docs.hetzner.cloud
const hetznerProvider string = "Hetzner"

type HCloud struct {
	client  *hcloud.Client
	context context.Context
}

func (cloud *HCloud) GetProvider() string {
	return hetznerProvider
}

func (cloud *HCloud) GetSSHKey(fingerprint string) (int, error) {
	sshKey, _, err := cloud.client.SSHKey.GetByFingerprint(cloud.context, fingerprint)
	if err != nil {
		return 0, newNotExistsError("ssh key", fingerprint, err)
	}

	return sshKey.ID, nil
}

func (cloud *HCloud) GetSnapshot(name string) (*Snapshot, error) {
	images, err := cloud.client.Image.All(cloud.context)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the image: %v", err)
	}

	lowerName := strings.ToLower(name)

	for _, image := range images {

		if image.Type != hcloud.ImageTypeSnapshot {
			continue
		}

		if strings.ToLower(image.Description) != lowerName {
			continue
		}

		return &Snapshot{
			Name: image.Description,
			Id:   image.ID,
		}, nil
	}

	return nil, newNotExistsError("image", name, nil)
}

func (cloud *HCloud) GetServer(name string) (*Server, error) {
	server, _, err := cloud.client.Server.GetByName(cloud.context, name)
	if err != nil {
		return nil, fmt.Errorf("couldn't list servers: %v", err)
	}

	if server == nil {
		return nil, newNotExistsError("server", name, nil)
	}

	return cloud.toCloudServer(server), nil
}

func (cloud *HCloud) StartServer(server *Server) error {
	action, _, err := cloud.client.Server.Poweron(cloud.context, &hcloud.Server{ID: server.Id})
	if err != nil {
		return fmt.Errorf("couldn't power on server: %v", err)
	}

	if action.Status == "errored" {
		return fmt.Errorf("shutdown with status 'errored' for server %v", server.Name)
	}

	return nil
}

func (cloud *HCloud) StopServer(server *Server) error {
	action, _, err := cloud.client.Server.Shutdown(cloud.context, &hcloud.Server{ID: server.Id})
	if err != nil {
		return fmt.Errorf("couldn't shutdown server: %v", err)
	}

	if action.Status == "errored" {
		return fmt.Errorf("shutdown with status 'errored' for server %v", server.Name)
	}

	return nil
}

func (cloud *HCloud) CreateServer(options CreateOptions) (*Server, error) {
	opts := hcloud.ServerCreateOpts{
		Name:       options.Name,
		ServerType: &hcloud.ServerType{Name: options.Machine},
		Location:   &hcloud.Location{Name: options.Region},
		Image:      &hcloud.Image{ID: options.Snapshot.Id},
		SSHKeys: []*hcloud.SSHKey{
			{ID: options.SshKey},
		},
	}
	result, _, err := cloud.client.Server.Create(cloud.context, opts)
	if err != nil {
		return nil, fmt.Errorf("couldn't create server: %v", err)
	}

	return cloud.toCloudServer(result.Server), nil
}

func (cloud *HCloud) DestroyServer(server *Server) error {
	_, err := cloud.client.Server.Delete(cloud.context, &hcloud.Server{ID: server.Id})
	if err != nil {
		return fmt.Errorf("couldn't delete server: %v", err)
	}

	server.Status = StatusDestroyed

	return nil
}

func (cloud *HCloud) toCloudServer(server *hcloud.Server) *Server {
	status := StatusOff

	switch server.Status {
	case hcloud.ServerStatusInitializing:
		status = StatusStartup
	case hcloud.ServerStatusRunning:
		status = StatusActive
	}

	return &Server{
		Name:     server.Name,
		Id:       server.ID,
		Ip:       server.PublicNet.IPv4.IP.String(),
		Provider: hetznerProvider,
		Status:   status,
	}
}

func newHCloud(token string) *HCloud {
	client := hcloud.NewClient(
		hcloud.WithToken(token),
		hcloud.WithApplication("StartMyGame", "v1"),
	)

	return &HCloud{
		client:  client,
		context: context.TODO(),
	}
}
