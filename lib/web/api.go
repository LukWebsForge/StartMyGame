package web

import (
	"encoding/json"
	"fmt"
	"github.com/rs/cors"
	"log"
	"net/http"
	"start-my-game/lib/cloud"
	"start-my-game/lib/config"
	"start-my-game/lib/manager"
	"strconv"
	"time"
)

type ApiServer struct {
	manger *manager.Manager
}

type StartResponse struct {
	// Can be 'already_running', 'in_startup', 'starting', 'creating' or 'failure'
	Status string `json:"status"`
}

type StatusResponse struct {
	// Can be 'active', 'startup', 'startup_error' or 'off'
	Status string `json:"status"`
	// Must be smaller or equal to ProgressMax
	Progress     int       `json:"progress"`
	ProgressMax  int       `json:"progress_max"`
	Ip           string    `json:"ip"`
	Name         string    `json:"name"`
	OnlinePlayer int       `json:"online_player"`
	LastOnline   time.Time `json:"last_online"`
}

func Start(cfg *config.Config, manager *manager.Manager) {
	api := ApiServer{
		manger: manager,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/start/", api.startHandler)
	mux.HandleFunc("/status/", api.statusHandler)

	handler := cors.New(cors.Options{
		AllowedOrigins: []string{cfg.Web.CorsDomain},
	}).Handler(mux)

	log.Printf("Starting web server on port %v\n", cfg.Web.Port)
	err := http.ListenAndServe(":"+strconv.Itoa(cfg.Web.Port), handler)

	if err != nil {
		log.Fatalf("couldn't start web server on port %v: %v\n", cfg.Web.Port, err)
	}
}

func (api *ApiServer) startHandler(writer http.ResponseWriter, request *http.Request) {
	// Only accepting POST requests
	if request.Method != "POST" {
		return
	}

	status := ""

	if api.manger.Startup != nil && api.manger.Startup.InProgress() {
		status = "in_startup"
	} else {
		api.manger.UpdateActiveServer()

		server := api.manger.ActiveServer
		if server == nil {
			log.Printf("Request which results in a start from %v", requestingAddr(request))
			go api.manger.CreateServer()
			status = "creating"
		} else {
			if server.Status == cloud.StatusActive {
				status = "already_running"
			} else {
				log.Printf("Request which results in a start from %v", requestingAddr(request))
				go api.manger.StartServer()
				status = "starting"
			}
		}
	}

	jsonResponse(writer, StartResponse{Status: status})
}

func requestingAddr(request *http.Request) string {
	forwarded := request.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded + " (" + request.RemoteAddr + ")"
	} else {
		return request.RemoteAddr
	}
}

func (api *ApiServer) statusHandler(writer http.ResponseWriter, request *http.Request) {
	jsonResponse(writer, generateStatusResponse(api.manger))
}

func generateStatusResponse(manager *manager.Manager) StatusResponse {
	server := manager.ActiveServer

	if manager.Startup != nil {
		// Returning the startup status
		startup := manager.Startup

		response := StatusResponse{
			Progress:     startup.Current,
			ProgressMax:  startup.Max,
			Ip:           "",
			Name:         "",
			OnlinePlayer: 0,
			LastOnline:   manager.LastActivePlayer,
		}

		if server != nil {
			response.Ip = server.Ip
		}

		if startup.InProgress() {
			response.Status = "startup"
			return response
		} else if startup.Error {
			response.Status = "startup_error"
			return response
		}
	}

	if server != nil {
		response := StatusResponse{
			Status:      manager.GetServerStatus(),
			Progress:    0,
			ProgressMax: 0,
			Ip:          server.Ip,
			LastOnline:  manager.LastActivePlayer,
		}

		// Handle the case if the application just was started
		if manager.LastRconInfo != nil {
			response.Name = manager.LastRconInfo.Name
			response.OnlinePlayer = manager.LastRconInfo.Online
		} else {
			response.Name = "Lädt..."
			response.OnlinePlayer = 0
		}

		return response
	} else {
		return StatusResponse{
			Status:       manager.GetServerStatus(),
			Progress:     0,
			ProgressMax:  0,
			Name:         "",
			Ip:           "",
			OnlinePlayer: 0,
			LastOnline:   manager.LastActivePlayer,
		}
	}
}

func jsonResponse(writer http.ResponseWriter, response interface{}) {
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	err := json.NewEncoder(writer).Encode(response)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(writer, err.Error())
	}
}
