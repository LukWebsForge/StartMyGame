package cloud

import (
	"context"
	"fmt"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"strings"
)

// https://developers.digitalocean.com/documentation/v2/
const digitalOceanProvider string = "DigitalOcean"

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

type DoCloud struct {
	client  *godo.Client
	context context.Context
}

func (cloud *DoCloud) GetProvider() string {
	return digitalOceanProvider
}

func (cloud *DoCloud) GetSSHKey(fingerprint string) (int, error) {
	key, _, err := cloud.client.Keys.GetByFingerprint(cloud.context, fingerprint)
	if err != nil {
		return 0, newNotExistsError("ssh key", fingerprint, err)
	}

	return key.ID, nil
}

func (cloud *DoCloud) GetSnapshot(name string) (*Snapshot, error) {
	listOptions := godo.ListOptions{
		Page:    1,
		PerPage: 200,
	}

	images, _, err := cloud.client.Images.ListUser(cloud.context, &listOptions)
	if err != nil {
		return nil, fmt.Errorf("couldn't list images: %v", err)
	}

	lowerName := strings.ToLower(strings.TrimRight(name, " "))

	for _, image := range images {

		if strings.Index(strings.ToLower(image.Name), lowerName) < 0 {
			continue
		}

		return &Snapshot{
			Name: image.Name,
			Id:   image.ID,
		}, nil
	}

	return nil, newNotExistsError("snapshot", name, nil)
}

func (cloud *DoCloud) GetServer(name string) (*Server, error) {
	listOptions := godo.ListOptions{
		Page:    1,
		PerPage: 200,
	}

	droplets, _, err := cloud.client.Droplets.List(cloud.context, &listOptions)
	if err != nil {
		return nil, fmt.Errorf("couldn't list droplets: %v", err)
	}

	lowerName := strings.Trim(strings.ToLower(name), "")

	for _, droplet := range droplets {

		if strings.Trim(strings.ToLower(droplet.Name), "") != lowerName {
			continue
		}

		ipv4, err := droplet.PublicIPv4()
		if err != nil {
			return nil, fmt.Errorf("couldn't get the Ip of the droplet: %v", err)
		}

		return cloud.dropletToServer(&droplet, ipv4)
	}

	return nil, newNotExistsError("server", name, nil)
}

func (cloud *DoCloud) StartServer(server *Server) error {
	action, _, err := cloud.client.DropletActions.PowerOn(cloud.context, server.Id)
	if err != nil {
		return fmt.Errorf("couldn't shutdown a droplet: %v", err)
	}

	if action.Status == "errored" {
		return fmt.Errorf("shutdown with status 'errored' for droplet %v", server.Name)
	}

	return nil
}

func (cloud *DoCloud) StopServer(server *Server) error {
	action, _, err := cloud.client.DropletActions.Shutdown(cloud.context, server.Id)
	if err != nil {
		return fmt.Errorf("couldn't shutdown a droplet: %v", err)
	}

	if action.Status == "errored" {
		return fmt.Errorf("shutdown with status 'errored' for droplet %v", server.Name)
	}

	return nil
}

func (cloud *DoCloud) CreateServer(options CreateOptions) (*Server, error) {
	request := &godo.DropletCreateRequest{
		Name:   options.Name,
		Region: options.Region,
		Size:   options.Machine,
		Image: godo.DropletCreateImage{
			ID: options.Snapshot.Id,
		},
		IPv6: true,
		SSHKeys: []godo.DropletCreateSSHKey{
			{ID: options.SshKey},
		},
		Monitoring: true,
	}

	droplet, _, err := cloud.client.Droplets.Create(cloud.context, request)
	if err != nil {
		return nil, fmt.Errorf("couldn't create droplet: %v", err)
	}

	ipv4 := droplet.Networks.V4[0].IPAddress
	return cloud.dropletToServer(droplet, ipv4)
}

func (cloud *DoCloud) DestroyServer(server *Server) error {
	_, err := cloud.client.Droplets.Delete(cloud.context, server.Id)
	if err != nil {
		return fmt.Errorf("couldn't delete the droplet: %v", err)
	}

	return nil
}

func (cloud *DoCloud) dropletToServer(droplet *godo.Droplet, ipv4 string) (*Server, error) {
	status := StatusOff

	switch droplet.Status {
	case "new":
		status = StatusStartup
	case "active":
		status = StatusActive
	}

	return &Server{
		Name:     droplet.Name,
		Id:       droplet.ID,
		Ip:       ipv4,
		Provider: digitalOceanProvider,
		Status:   status,
	}, nil
}

func newDoCloud(token string) *DoCloud {
	tokenSource := &TokenSource{
		AccessToken: token,
	}

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	client := godo.NewClient(oauthClient)

	return &DoCloud{
		client:  client,
		context: context.TODO(),
	}
}
