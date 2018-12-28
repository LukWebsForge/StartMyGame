package manager

import (
	"fmt"
	"log"
	"start-my-game/lib/cloud"
	"start-my-game/lib/gmod"
	"time"
)

type StartupProgress struct {
	start   time.Time
	Current int
	Max     int
	Error   bool
}

func (progress *StartupProgress) InProgress() bool {
	return progress.Current < progress.Max && !progress.Error
}

func (manager *Manager) UpdateActiveServer() {
	server, err := manager.cloud.GetServer(manager.config.Cloud.ServerName)

	if err != nil {
		if !cloud.IsNotExistsError(err) {
			fmt.Println("Error during active server update:", err)
		}
		manager.ActiveServer = nil
	} else {
		manager.ActiveServer = server
	}
}

// UpdateActiveServer should be called before running this method
func (manager *Manager) CreateServer() {

	manager.Startup = &StartupProgress{
		start:   time.Now(),
		Current: 0,
		Max:     5,
		Error:   false,
	}

	if manager.ActiveServer != nil && manager.ActiveServer.Status != cloud.StatusDestroyed {
		startupError(manager, fmt.Errorf("won't create a new server because there's a server with status %v",
			manager.ActiveServer.Status))
		return
	}

	log.Printf("Starting to create a new server...\n")

	// Get the SSH key id
	key, err := manager.cloud.GetSSHKey(manager.config.Cloud.SshKey)
	if err != nil {
		startupError(manager, err)
		return
	}

	startupNext(manager)

	// Get the snapshot id
	snapshot, err := manager.cloud.GetSnapshot(manager.config.Cloud.Snapshot)
	if err != nil {
		startupError(manager, err)
		return
	}

	startupNext(manager)

	// Creating the server
	log.Printf("Creating a new server with\n ssh key: '%v'\n snapshot '%v'\n machine '%v'\n region '%v'\n",
		manager.config.Cloud.SshKey, snapshot.Name, manager.config.Cloud.ServerType, manager.config.Cloud.Region)

	server, err := manager.cloud.CreateServer(cloud.CreateOptions{
		Name:     manager.config.Cloud.ServerName,
		Machine:  manager.config.Cloud.ServerType,
		Region:   manager.config.Cloud.Region,
		Snapshot: snapshot,
		SshKey:   key,
	})
	if err != nil {
		startupError(manager, err)
		return
	}

	startupNext(manager)
	log.Printf("Server '%v' got the IP %v\n", server.Name, server.Ip)

	serverStartupCheck(manager, server)
}

// UpdateActiveServer should be called before running this method
func (manager *Manager) StartServer() {

	server := manager.ActiveServer
	manager.Startup = &StartupProgress{
		start:   time.Now(),
		Current: 0,
		Max:     3,
		Error:   false,
	}

	if server == nil {
		startupError(manager, fmt.Errorf("can't start a non existing server"))
		return
	}

	if server.Status != cloud.StatusOff {
		startupError(manager, fmt.Errorf("can't start a server with state %v", server.Status))
		return
	}

	err := manager.cloud.StartServer(server)
	if err != nil {
		startupError(manager, err)
		return
	}

	startupNext(manager)
	serverStartupCheck(manager, server)
}

func serverStartupCheck(manager *Manager, server *cloud.Server) {
	// Waiting 5 minutes for server boot; Checking every 30 seconds
	online := false
	for i := 0; i < 10; i++ {

		server, err := manager.cloud.GetServer(manager.config.Cloud.ServerName)
		if err != nil {
			log.Println("Error while server boot check:", err)
			continue
		}

		if server.Status != cloud.StatusActive {
			time.Sleep(30 * time.Second)
			continue
		}

		online = true
		break
	}

	if !online {
		startupError(manager, fmt.Errorf("server not online after 5 minutes"))
		return
	}

	log.Printf("Server '%v' is online, waiting for gmod...\n", server.Name)

	manager.ActiveServer = server
	startupNext(manager)

	// Waiting 5 minutes for the gmod server start
	online = false
	for i := 0; i < 20; i++ {
		_, err := gmod.NewRcon(server.Ip, manager.config)
		if err != nil {
			time.Sleep(15 * time.Second)
			continue
		}

		online = true
	}

	if !online {
		startupError(manager, fmt.Errorf("gmod rcon not responding after 5 mintues"))
		return
	}

	log.Printf("Gmod is online, everything was successful!")

	startupNext(manager)
	manager.UpdateActiveServer()
	manager.LastActivePlayer = time.Now()
}

func startupError(manager *Manager, err error) {
	manager.Startup.Error = true
	log.Println("Error while server startup:", err)
}

func startupNext(manager *Manager) {
	manager.Startup.Current++
}

// UpdateActiveServer should be called before running this method
func (manager *Manager) deleteServer() {
	server := manager.ActiveServer
	manager.ActiveServer = nil

	if server.Status == cloud.StatusDestroyed {
		log.Println("Won't delete a destroyed server")
		return
	}

	// Gracefully stopping the server if online
	if server.Status == cloud.StatusActive {
		err := manager.cloud.StopServer(server)
		if err != nil {
			log.Println("Couldn't stop server", err)
		}

		log.Println("Stopping the server", server.Name)

		time.Sleep(time.Duration(30 * time.Second))
	}

	log.Printf("Destroying server %v...\n", server.Name)
	// Deleting the virtual server instance
	err := manager.cloud.DestroyServer(server)
	if err != nil {
		log.Panicln("Couldn't destroy server:", err)
	}

	server.Status = cloud.StatusDestroyed

	log.Println("Destroyed server", server.Name)
}
