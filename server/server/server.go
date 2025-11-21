package server

import (
	"encoding/json"
	"flag"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

var Addr = flag.String("addr", "0.0.0.0:8080", "http service address")
var upgrader = websocket.Upgrader{} // use default options

var sessionStorage = MakeSessionStorage()

var runningSessionsGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "running_game_sessions",
	Help: "The total number of currently running game sessions",
})
var createdSessionsCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "game_sessions_created",
	Help: "The total number of created game sessions",
})

type CreateSessionResponse struct {
	SessionId int64 `json:"sessionId"`
}

type SessionDto struct {
	SessionId int64 `json:"sessionId"`
	Started   bool  `json:"started"`
}

func GetSessionsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionDtos := make([]SessionDto, 0)
	for _, session := range sessionStorage.GetAll() {
		sessionDtos = append(sessionDtos, SessionDto{SessionId: session.sessionId, Started: session.started.Load()})
	}
	json.NewEncoder(w).Encode(sessionDtos)
}

func CreateSession(w http.ResponseWriter, r *http.Request) {
	gameSession := makeGameSession()
	sessionStorage.Add(gameSession)
	createdSessionsCounter.Inc()
	log.Infof("Session %d created", gameSession.sessionId)
	response := CreateSessionResponse{SessionId: gameSession.sessionId}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ConnectToSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// TODO: sessionStorage
	sessionId, _ := strconv.ParseInt(vars["sessionId"], 10, 64)
	gameSession, _ := sessionStorage.Get(sessionId)

	if gameSession.started.Load() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Session already started"))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warn("upgrade:", err)
		return
	}

	gameSession.addPlayer(conn)
}
