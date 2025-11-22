package server

import (
	"encoding/json"
	"flag"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

var Addr = flag.String("addr", "0.0.0.0:8080", "http service address")
var upgrader = websocket.Upgrader{} // use default options

var sessionStorage = MakeSessionStorage()
var minVersion = "3.0.0"

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
	clientVersion := r.Header.Get("X-Client-Version")
	if clientVersion == "" {
		clientVersion = "0.0.0"
	}

	if compareVersions(clientVersion, minVersion) < 0 {
		w.WriteHeader(http.StatusUpgradeRequired)
		json.NewEncoder(w).Encode(map[string]string{
			"error":        "client_version_too_old",
			"required_min": minVersion,
			"your_version": clientVersion,
			"message":      "Please update your client to continue",
		})
		return
	}

	sessionDtos := make([]SessionDto, 0)
	for _, session := range sessionStorage.GetAll() {
		sessionDtos = append(sessionDtos, SessionDto{SessionId: session.sessionId, Started: session.started.Load()})
	}
	json.NewEncoder(w).Encode(sessionDtos)
}

func compareVersions(v1, v2 string) int {
	s1 := strings.Split(v1, ".")
	s2 := strings.Split(v2, ".")

	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	// сравниваем по сегментам
	for i := 0; i < maxLen; i++ {
		var n1, n2 int

		if i < len(s1) {
			n1, _ = strconv.Atoi(s1[i])
		}

		if i < len(s2) {
			n2, _ = strconv.Atoi(s2[i])
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
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
