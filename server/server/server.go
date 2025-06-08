package server

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jellydator/ttlcache/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var Addr = flag.String("addr", "0.0.0.0:8080", "http service address")
var upgrader = websocket.Upgrader{} // use default options

var Sessions = make(map[int64]*GameSession)
var PlayersPingMeasurer = MakePingMeasurer()

var runningSessionsGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "running_game_sessions",
	Help: "The total number of currently running game sessions",
})
var createdSessionsCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "game_sessions_created",
	Help: "The total number of created game sessions",
})
var pingHist = promauto.NewHistogram(prometheus.HistogramOpts{
	Name:    "ping",
	Help:    "Ping ms distribution histogram",
	Buckets: []float64{50, 100, 250, 500, 1000},
})

type PingMeasurer struct {
	pingMeasures *ttlcache.Cache[uuid.UUID, time.Time]
}

func MakePingMeasurer() *PingMeasurer {
	cache := ttlcache.New[uuid.UUID, time.Time](
		ttlcache.WithTTL[uuid.UUID, time.Time](time.Minute),
	)
	return &PingMeasurer{
		pingMeasures: cache,
	}
}

func (pingMeasurer *PingMeasurer) addMeasure() uuid.UUID {
	pingUuid := uuid.New()
	pingMeasurer.pingMeasures.Set(pingUuid, time.Now(), ttlcache.DefaultTTL)
	return pingUuid
}

func (pingMeasurer *PingMeasurer) getMeasure(uuid uuid.UUID) (time.Time, bool) {
	startTime := pingMeasurer.pingMeasures.Get(uuid)
	if startTime != nil {
		return startTime.Value(), true
	} else {
		return time.Time{}, false
	}
}

func (pingMeasurer *PingMeasurer) getMeasuresCount() int {
	return pingMeasurer.pingMeasures.Len()
}

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
	for sessionId := range Sessions {
		session := Sessions[sessionId]
		sessionDtos = append(sessionDtos, SessionDto{SessionId: sessionId, Started: session.Started})
	}
	json.NewEncoder(w).Encode(sessionDtos)
}

func CreateSession(w http.ResponseWriter, r *http.Request) {
	gameSession := MakeGameSession()
	Sessions[gameSession.GetSessionId()] = gameSession
	response := CreateSessionResponse{SessionId: gameSession.GetSessionId()}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ConnectToSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionId, _ := strconv.ParseInt(vars["sessionId"], 10, 64)
	session := Sessions[sessionId]

	if session.Started {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Session already started"))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warn("upgrade:", err)
		return
	}

	if session.FirstPlayerSession == nil {
		firstPlayerPieceGenerator := rand.New(rand.NewSource(sessionId))
		firstPlayerSession := MakePlayerSession(conn, firstPlayerPieceGenerator, session)
		session.FirstPlayerSession = firstPlayerSession
		log.Infof("Session %d created", sessionId)
		createdSessionsCounter.Inc()
	} else {
		firstPlayerSession := session.FirstPlayerSession
		secondPlayerPieceGenerator := rand.New(rand.NewSource(sessionId))
		secondPlayerSession := MakePlayerSession(conn, secondPlayerPieceGenerator, session)
		session.SecondPlayerSession = secondPlayerSession
		firstPlayerSession.EnemySession = secondPlayerSession
		secondPlayerSession.EnemySession = firstPlayerSession
		firstPlayerSession.conn.SetPongHandler(pongHandler(firstPlayerSession))
		secondPlayerSession.conn.SetPongHandler(pongHandler(secondPlayerSession))
		session.Started = true
		session.RunSession()
		runningSessionsGauge.Inc()
		log.Infof("Session %d started", sessionId)
	}
}

func pongHandler(playerSession *PlayerSession) func(appData string) error {
	return func(appData string) error {
		pingUuid, _ := uuid.FromBytes([]byte(appData))
		if startTime, exists := PlayersPingMeasurer.getMeasure(pingUuid); exists {
			ping := time.Since(startTime).Milliseconds()
			pingHist.Observe(float64(ping))
			message := fmt.Sprintf("%d %d", 2, ping)
			playerSession.SendMessage(message)
			return nil
		} else {
			log.Warn("Ping UUID not found", pingUuid)
			return nil
		}
	}
}
