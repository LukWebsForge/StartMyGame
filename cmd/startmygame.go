package main

import (
	"log"
	"os"
	"start-my-game/lib/cloud"
	"start-my-game/lib/config"
	"start-my-game/lib/manager"
	"start-my-game/lib/web"
)

func main() {
	// Create default config
	created, err := config.CreateIfNotExists()
	if err != nil {
		log.Panicln("Couldn't create config:", err)
	}

	if created == true {
		log.Println("A config file has been created. Edit it and start the application again")
		os.Exit(2)
	}

	// Read config
	cfg, err := config.Read()
	if err != nil {
		log.Panicln("Couldn't read config:", err)
	}

	// Create cloud
	acloud, err := cloud.NewCloud(cfg)
	if err != nil {
		log.Panicln("Couldn't init cloud:", err)
	}

	log.Printf("Initalized cloud with provider %v\n", acloud.GetProvider())

	newManager := manager.NewManager(cfg, acloud)
	// go newManager.DelayCheckStart()
	go newManager.StartCheck()

	// TODO: Run with go
	web.Start(cfg, newManager)
}
