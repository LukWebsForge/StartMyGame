package manager

import (
	"log"
	"start-my-game/lib/cloud"
	"start-my-game/lib/config"
	"start-my-game/lib/gmod"
	"time"
)

type Manager struct {
	LastActivePlayer time.Time
	LastRconInfo     *gmod.GServerInfo
	ActiveServer     *cloud.Server
	Startup          *StartupProgress
	config           *config.Config
	cloud            cloud.Cloud
}

func (manager *Manager) interval() time.Duration {
	return time.Duration(manager.config.Gmod.CheckInterval) * time.Minute
}

func (manager *Manager) shutdownDelay() time.Duration {
	return time.Duration(manager.config.Gmod.ShutdownAfter) * time.Minute
}

func NewManager(cfg *config.Config, acloud cloud.Cloud) *Manager {
	manager := Manager{
		LastActivePlayer: time.Time{},
		LastRconInfo:     nil,
		ActiveServer:     nil,
		config:           cfg,
		cloud:            acloud,
	}

	manager.LastActivePlayer = time.Now().Add(manager.shutdownDelay() / -2)
	manager.UpdateActiveServer()

	return &manager
}

func (manager *Manager) DelayCheckStart() {
	fullInterval := time.Now().Truncate(manager.interval()).Add(manager.interval())

	log.Printf("Full interval at %v\n", fullInterval)
	time.Sleep(time.Until(fullInterval))

	manager.StartCheck()
}

func (manager *Manager) StartCheck() {
	manager.UpdateActiveServer()
	if manager.ActiveServer == nil {
		log.Println("At the beginning there was nothing")
	} else {
		server := manager.ActiveServer
		log.Printf("At the begining there was a server with the name %v, the ip %v. It was very %v.\n",
			server.Name, server.Ip, server.Status)

		if manager.ActiveServer.Status == cloud.StatusStartup {
			// Showing a startup bar, if the app and the server are starting
			manager.Startup = &StartupProgress{
				start:   time.Now(),
				Current: 3,
				Max:     5,
				Error:   false,
			}
			manager.ActiveServer = nil
			serverStartupCheck(manager, server)
		}
	}

	timer := time.NewTicker(manager.interval())
	log.Printf("Online check started, running every %v\n", manager.interval())
	manager.check()

	for range timer.C {
		manager.check()
	}
}

func (manager *Manager) check() {
	if manager.ActiveServer == nil {
		return
	}

	if manager.Startup != nil && manager.Startup.InProgress() {
		return
	}

	if manager.ActiveServer.Status == cloud.StatusActive {
		rcon, err := gmod.NewRcon(manager.ActiveServer.Ip, manager.config)
		if err != nil {
			log.Println("Couldn't connect via rcon to the server:", err)
			return
		}

		rconInfo, err := rcon.ServerStatus()
		if err != nil {
			log.Println("Couldn't read online players via rcon", err)
			return
		}

		manager.LastRconInfo = rconInfo

		if rconInfo.Online > 0 {
			log.Printf("%v of %v players online\n", rconInfo.Online, rconInfo.Max)
			manager.LastActivePlayer = time.Now()
			return
		}
	}

	emptyDuration := time.Since(manager.LastActivePlayer)
	// log.Printf("Empty Duration: %v ShutdownDelay: %v", emptyDuration.Seconds(), manager.shutdownDelay().Seconds())
	if emptyDuration.Seconds() >= manager.shutdownDelay().Seconds() {
		manager.deleteServer()
	}
}
