package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const configPath string = "config.json"

type Config struct {
	Web   Web   `json:"web"`
	Gmod  Gmod  `json:"gmod"`
	Cloud Cloud `json:"cloud"`
}

type Web struct {
	Port       int    `json:"port"`
	CorsDomain string `json:"cors_domain"`
}

type Gmod struct {
	Port          int    `json:"port"`
	Password      string `json:"rcon_password"`
	CheckInterval int    `json:"check_interval"`
	ShutdownAfter int    `json:"shutdown_after"`
}

type Cloud struct {
	Provider   string `json:"provider"`
	Token      string `json:"token"`
	ServerName string `json:"server_name"`
	ServerType string `json:"server_type"`
	Region     string `json:"region"`
	Snapshot   string `json:"snapshot"`
	SshKey     string `json:"ssh_key"`
}

func Read() (*Config, error) {
	conf := Config{}

	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("can't load config: %v", err)
	}

	err = json.Unmarshal(bytes, &conf)
	if err != nil {
		return nil, fmt.Errorf("can't read config: %v", err)
	}

	return &conf, nil
}

func CreateIfNotExists() (bool, error) {
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		return false, nil
	}

	defaultConf := Config{
		Cloud: Cloud{
			Provider:   "Hetzner",
			Token:      "YourCloudToken",
			ServerName: "YourDropletOrServerName",
			ServerType: "TheCloudServerType",
			Region:     "TheCloudRegion",
			Snapshot:   "YourSnapshotName",
			SshKey:     "YourSshKeyFingerprint",
		},
		Gmod: Gmod{
			Password:      "YourRconPassword",
			Port:          27015,
			CheckInterval: 5,
			ShutdownAfter: 60,
		},
		Web: Web{Port: 8011, CorsDomain: "http://ttt.example.com"},
	}

	bytes, err := json.MarshalIndent(defaultConf, "", "    ")
	if err != nil {
		return false, fmt.Errorf("can't compose config: %v", err)
	}

	// Permission 0640: The user can read and write, the group can read
	err = ioutil.WriteFile(configPath, bytes, os.FileMode(0640))
	if err != nil {
		return false, fmt.Errorf("can't write config: %v", err)
	}

	return true, nil
}
