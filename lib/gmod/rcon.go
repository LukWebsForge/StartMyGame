package gmod

import (
	"fmt"
	"github.com/james4k/rcon"
	"regexp"
	"start-my-game/lib/config"
	"strconv"
)

type Rcon struct {
	rcon *rcon.RemoteConsole
}

type GServerInfo struct {
	Name   string
	Online int
	Max    int
}

func (gmod *Rcon) ServerStatus() (*GServerInfo, error) {

	requestId, err := gmod.rcon.Write("status")
	if err != nil {
		return nil, fmt.Errorf("couldn't send request for online players: %v", err)
	}

	response, responseId, err := gmod.rcon.Read()
	if err != nil {
		return nil, fmt.Errorf("coudln't read response for online players: %v", err)
	} else if requestId != responseId {
		return nil, fmt.Errorf("couldn't read response for online players, becuase of invalid answer id")
	}

	online, max, err := extractPlayerCount(response)
	if err != nil {
		return nil, fmt.Errorf("online player count: %v", err)
	}
	name, err := extractServerName(response)
	if err != nil {
		return nil, fmt.Errorf("hostname: %v", err)
	}

	return &GServerInfo{Name: name, Online: online, Max: max}, nil
}

func extractPlayerCount(response string) (int, int, error) {
	compile, err := regexp.Compile(`players : (\d+) \((\d+) max\)`)

	if err != nil {
		return 0, 0, fmt.Errorf("couldn't compile regex: %v", err)
	}

	submatch := compile.FindAllStringSubmatch(response, 1)
	if len(submatch) == 0 {
		return 0, 0, fmt.Errorf("couldn't find a match")
	}

	online, errOnline := strconv.Atoi(submatch[0][1])
	max, errMax := strconv.Atoi(submatch[0][2])
	if errOnline != nil || errMax != nil {
		err = fmt.Errorf("couldn't convert response match to ints: '%v'", submatch[0][0])
	}

	return online, max, nil
}

func extractServerName(response string) (string, error) {
	compile, err := regexp.Compile(`hostname: ([A-Za-z0-9 ]+)`)

	if err != nil {
		return "", fmt.Errorf("couldn't compile regex: %v", err)
	}

	submatch := compile.FindAllStringSubmatch(response, 1)
	if len(submatch) == 0 {
		return "", fmt.Errorf("couldn't find a match")
	}

	return submatch[0][1], nil
}

func NewRcon(ip string, config *config.Config) (*Rcon, error) {
	remoteAddr := fmt.Sprintf("%v:%v", ip, config.Gmod.Port)
	console, err := rcon.Dial(remoteAddr, config.Gmod.Password)
	if err != nil {
		return nil, fmt.Errorf("couldn't connect via rcon to '%v': %v", remoteAddr, err)
	}

	return &Rcon{
		rcon: console,
	}, nil
}
