package server

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"tetrisServer/field"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

type GameSession struct {
	sessionId           int64
	firstPlayerSession  *PlayerSession
	secondPlayerSession *PlayerSession
	started             atomic.Bool
	someEnded           atomic.Bool
}

func makeGameSession() *GameSession {
	sessionId := time.Now().Unix()
	return &GameSession{
		sessionId: sessionId,
	}
}

func (gameSession *GameSession) addPlayer(playerConnection *websocket.Conn) {
	pieceGenerator := rand.New(rand.NewSource(gameSession.sessionId))
	playerSession := makePlayerSession(playerConnection, pieceGenerator, gameSession)
	if gameSession.firstPlayerSession == nil {
		gameSession.firstPlayerSession = playerSession
	} else {
		gameSession.secondPlayerSession = playerSession
		gameSession.firstPlayerSession.enemySession = gameSession.secondPlayerSession
		gameSession.secondPlayerSession.enemySession = gameSession.firstPlayerSession
		gameSession.firstPlayerSession.conn.SetPongHandler(pongHandler(gameSession.firstPlayerSession))
		gameSession.secondPlayerSession.conn.SetPongHandler(pongHandler(gameSession.secondPlayerSession))
		gameSession.startSession()
	}
}

func (gameSession *GameSession) startSession() {
	gameSession.started.Store(true)
	gameSession.firstPlayerSession.startSession()
	gameSession.secondPlayerSession.startSession()
	runningSessionsGauge.Inc()
	log.Infof("Session %d started", gameSession.sessionId)
}

func (gameSession *GameSession) endSession() {
	firstPlayerScore := gameSession.firstPlayerSession.playerField.GetScore()
	secondPlayerScore := gameSession.secondPlayerSession.playerField.GetScore()

	if firstPlayerScore > secondPlayerScore {
		gameSession.firstPlayerSession.sendMessage("0 0 WIN!")
		gameSession.secondPlayerSession.sendMessage("0 0 LOSE(")
	} else if secondPlayerScore > firstPlayerScore {
		gameSession.firstPlayerSession.sendMessage("0 0 LOSE(")
		gameSession.secondPlayerSession.sendMessage("0 0 WIN!")
	} else {
		gameSession.firstPlayerSession.sendMessage("0 0 DRAW=")
		gameSession.secondPlayerSession.sendMessage("0 0 DRAW=")
	}

	gameSession.firstPlayerSession.conn.Close()
	gameSession.secondPlayerSession.conn.Close()
	delete(Sessions, gameSession.sessionId)
	runningSessionsGauge.Dec()
	log.Infof("Session %d ended", gameSession.sessionId)
}

type WSConn interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	SetWriteDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	Close() error
	SetPongHandler(func(appData string) error)
}

type PlayerSession struct {
	playerField        field.Field
	conn               WSConn
	playerInputChannel chan rune
	isEnded            atomic.Bool
	pieceGenerator     *rand.Rand
	enemySession       *PlayerSession
	mu                 sync.Mutex
	gameSession        *GameSession
}

func makePlayerSession(conn *websocket.Conn, pieceGenerator *rand.Rand, gameSession *GameSession) *PlayerSession {
	pieceSelector := field.MakePieceSelector(pieceGenerator)
	field := field.MakeDefaultField(pieceSelector)
	session := PlayerSession{
		playerField:        field,
		conn:               conn,
		playerInputChannel: make(chan rune),
		pieceGenerator:     pieceGenerator,
		gameSession:        gameSession,
	}
	return &session
}

// sendMessage thread safe socket text message sending
func (playerSession *PlayerSession) sendMessage(message string) {
	playerSession.mu.Lock()
	defer playerSession.mu.Unlock()
	playerSession.conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// sendPingMessage thread safe socket ping message sending
func (playerSession *PlayerSession) sendPingMessage(pingUuid uuid.UUID) error {
	playerSession.mu.Lock()
	defer playerSession.mu.Unlock()
	pingUuidBinary, _ := pingUuid.MarshalBinary()
	return playerSession.conn.WriteMessage(websocket.PingMessage, pingUuidBinary)
}

func (playerSession *PlayerSession) startSession() {
	go playerSession.processPlayerInput()
	go playerSession.processGameField()
	go playerSession.processPlayerPing()
}

func (playerSession *PlayerSession) processPlayerPing() {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			playerSession.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			pingUuid := PlayersPingMeasurer.addMeasure()
			if err := playerSession.sendPingMessage(pingUuid); err != nil {
				return
			}
		}
	}
}

func (playerSession *PlayerSession) processGameField() {
	gameField := playerSession.playerField
	for {
		playerSession.inputControl()

		if !gameField.MovePiece(field.PieceMoveDown) {
			gameField.ConcatPiece()
			gameField.SelectNextPiece()
			if !gameField.CanMovePiece(field.PieceMoveDown) {
				playerSession.endSession()
				break
			}
			gameField.CleanLines()
		}
	}
}

// ends player session. ends game session if enemy session already ended
func (playerSession *PlayerSession) endSession() {
	playerSession.isEnded.Store(true)
	if playerSession.gameSession.someEnded.CompareAndSwap(false, true) {
		gameField := playerSession.playerField
		// add last piece to field to not lose it
		gameField.ConcatPiece()
		playerSession.sendMessage(formatFieldMessage(0, 1, gameField))
		playerSession.enemySession.sendMessage(formatFieldMessage(1, 1, gameField))
	} else {
		playerSession.gameSession.endSession()
	}
}

func (playerSession *PlayerSession) processPlayerInput() {
	for !playerSession.isEnded.Load() {
		_, message, err := playerSession.conn.ReadMessage()
		if err != nil {
			break
		}
		decodeRune, _ := utf8.DecodeRune(message)
		playerSession.playerInputChannel <- decodeRune
	}
}

func (playerSession *PlayerSession) inputControl() {
	gameField := playerSession.playerField
	timeout := time.After(time.Second / 4 / time.Duration(gameField.GetSpeed()))
	for {
		playerSession.sendMessage(formatFieldMessage(0, 1, gameField))
		playerSession.enemySession.sendMessage(formatFieldMessage(1, 1, gameField))
		select {
		case moveType := <-playerSession.playerInputChannel:
			switch moveType {
			case 'd':
				gameField.MovePiece(field.PieceMoveLeft)
			case 'a':
				gameField.MovePiece(field.PieceMoveRight)
			case 's':
				gameField.MovePiece(field.PieceMoveDown)
			case 'q':
				gameField.RotatePiece(field.PieceRotateLeft)
			case 'e':
				gameField.RotatePiece(field.PieceRotateRight)
			}
		case <-timeout:
			return
		}
	}
}

func formatFieldMessage(isEnemyField int, isAlive int, gameField field.Field) string {
	return fmt.Sprintf("%d %d %s %d %d %d %d", isEnemyField, isAlive, gameField.String(), gameField.GetSpeed(), gameField.GetScore(), gameField.GetCleanCount(), gameField.GetNextPieceType())
}
