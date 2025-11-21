package server

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"sync"
	"sync/atomic"
	"tetrisServer/field"
	"time"
	"unicode/utf8"

	log "github.com/sirupsen/logrus"

	lib "github.com/vszholobov/tetrisLib"

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
		gameSession.firstPlayerSession.conn.SetPingHandler(func(pingId string) error {
			return gameSession.firstPlayerSession.sendPongMessage([]byte(pingId))
		})
		gameSession.secondPlayerSession.conn.SetPingHandler(func(pingId string) error {
			return gameSession.secondPlayerSession.sendPongMessage([]byte(pingId))
		})
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
		gameSession.firstPlayerSession.sendMessage(lib.Self, lib.Win, gameSession.firstPlayerSession.playerField)
		gameSession.secondPlayerSession.sendMessage(lib.Self, lib.Lose, gameSession.secondPlayerSession.playerField)
	} else if secondPlayerScore > firstPlayerScore {
		gameSession.firstPlayerSession.sendMessage(lib.Self, lib.Lose, gameSession.firstPlayerSession.playerField)
		gameSession.secondPlayerSession.sendMessage(lib.Self, lib.Win, gameSession.secondPlayerSession.playerField)
	} else {
		gameSession.firstPlayerSession.sendMessage(lib.Self, lib.Draw, gameSession.firstPlayerSession.playerField)
		gameSession.secondPlayerSession.sendMessage(lib.Self, lib.Draw, gameSession.secondPlayerSession.playerField)
	}

	gameSession.firstPlayerSession.conn.Close()
	gameSession.secondPlayerSession.conn.Close()
	sessionStorage.Delete(gameSession.sessionId)
	runningSessionsGauge.Dec()
	log.Infof("Session %d ended", gameSession.sessionId)
}

type WSConn interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	SetWriteDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	Close() error
	SetPingHandler(func(appData string) error)
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
func (playerSession *PlayerSession) sendMessage(fieldType lib.FieldType, gameResult lib.GameResult, gameField field.Field) {
	playerSession.mu.Lock()
	defer playerSession.mu.Unlock()

	fieldBytes := gameField.Bytes()
	message := lib.GameStateMessage{
		FieldType:  fieldType,
		GameResult: gameResult,
		Speed:      uint8(gameField.GetSpeed()),
		Score:      uint16(gameField.GetScore()),
		CleanCount: uint16(gameField.GetCleanCount()),
		NextPiece:  lib.PieceType(gameField.GetNextPieceType()),
	}
	copy(message.FieldBytes[:], fieldBytes[:])

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, message); err != nil {
		log.Error(err.Error())
		return
	}
	playerSession.conn.WriteMessage(websocket.TextMessage, buf.Bytes())
}

// sendPingMessage thread safe socket ping message sending
func (playerSession *PlayerSession) sendPongMessage(pingId []byte) error {
	playerSession.mu.Lock()
	defer playerSession.mu.Unlock()
	return playerSession.conn.WriteMessage(websocket.PongMessage, pingId)
}

func (playerSession *PlayerSession) startSession() {
	go playerSession.processPlayerInput()
	go playerSession.processGameField()
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
		playerSession.sendMessage(lib.Self, lib.Ongoing, gameField)
		playerSession.enemySession.sendMessage(lib.Enemy, lib.Ongoing, gameField)
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
		playerSession.sendMessage(lib.Self, lib.Ongoing, gameField)
		playerSession.enemySession.sendMessage(lib.Enemy, lib.Ongoing, gameField)
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
