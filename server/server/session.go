package server

import (
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"sync"
	"tetrisServer/field"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

type GameSession struct {
	sessionId           int64
	FirstPlayerSession  *PlayerSession
	SecondPlayerSession *PlayerSession
	Started             bool
}

type PlayerSession struct {
	playerField        *field.Field
	conn               *websocket.Conn
	playerInputChannel chan rune
	isEnded            bool
	pieceGenerator     *rand.Rand
	EnemySession       *PlayerSession
	mu                 sync.Mutex
	gameSession        *GameSession
}

func MakeGameSession() *GameSession {
	sessionId := time.Now().Unix()
	return &GameSession{
		sessionId: sessionId,
	}
}

func MakePlayerSession(conn *websocket.Conn, pieceGenerator *rand.Rand, gameSession *GameSession) *PlayerSession {
	field := field.MakeDefaultField(pieceGenerator)
	session := PlayerSession{
		playerField:        &field,
		conn:               conn,
		playerInputChannel: make(chan rune),
		isEnded:            false,
		pieceGenerator:     pieceGenerator,
		gameSession:        gameSession,
	}
	return &session
}

func (gameSession *GameSession) RunSession() {
	gameSession.FirstPlayerSession.RunSession()
	gameSession.SecondPlayerSession.RunSession()
}

func (gameSession *GameSession) GetSessionId() int64 {
	return gameSession.sessionId
}

// SendMessage thread safe socket text message sending
func (playerSession *PlayerSession) SendMessage(message string) {
	playerSession.mu.Lock()
	defer playerSession.mu.Unlock()
	playerSession.conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// SendPingMessage thread safe socket ping message sending
func (playerSession *PlayerSession) SendPingMessage(pingUuid uuid.UUID) error {
	playerSession.mu.Lock()
	defer playerSession.mu.Unlock()
	pingUuidBinary, _ := pingUuid.MarshalBinary()
	return playerSession.conn.WriteMessage(websocket.PingMessage, pingUuidBinary)
}

func (playerSession *PlayerSession) RunSession() {
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
			if err := playerSession.SendPingMessage(pingUuid); err != nil {
				return
			}
		}
	}
}

func (playerSession *PlayerSession) processGameField() {
	gameField := playerSession.playerField
	for {
		playerSession.inputControl()

		if !gameField.CurrentPiece.MoveDown() {
			gameField.Val.Or(gameField.Val, gameField.CurrentPiece.GetVal())
			gameField.SelectNextPiece()
			if !gameField.CurrentPiece.CanMoveDown() {
				playerSession.endSession(gameField)
				break
			}
			gameField.Clean()
		}
	}
}

func (playerSession *PlayerSession) endSession(gameField *field.Field) {
	// TODO: race
	playerSession.isEnded = true
	if playerSession.EnemySession.isEnded {
		playerScore := *playerSession.playerField.Score
		enemyScore := *playerSession.EnemySession.playerField.Score

		if playerScore > enemyScore {
			playerSession.SendMessage("0 0 WIN!")
			playerSession.EnemySession.SendMessage("0 0 LOSE(")
		} else if enemyScore > playerScore {
			playerSession.SendMessage("0 0 LOSE(")
			playerSession.EnemySession.SendMessage("0 0 WIN!")
		} else {
			playerSession.SendMessage("0 0 DRAW=")
			playerSession.EnemySession.SendMessage("0 0 DRAW=")
		}

		playerSession.EnemySession.conn.Close()
		playerSession.conn.Close()
		sessionId := playerSession.gameSession.sessionId
		delete(Sessions, sessionId)
		log.Infof("Session %d ended", sessionId)
		runningSessionsGauge.Dec()
	} else {
		// add last piece to field to not lose it
		gameField.Val.Or(gameField.Val, gameField.CurrentPiece.GetVal())
		playerSession.SendMessage(FormatFieldMessage(0, 1, gameField))
		playerSession.EnemySession.SendMessage(FormatFieldMessage(1, 1, gameField))
	}
}

func (playerSession *PlayerSession) processPlayerInput() {
	for !playerSession.isEnded {
		// TODO: ticker чтобы не зависнуть когда сессия закончилась
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
		playerSession.SendMessage(FormatFieldMessage(0, 1, gameField))
		playerSession.EnemySession.SendMessage(FormatFieldMessage(1, 1, gameField))
		select {
		case moveType := <-playerSession.playerInputChannel:
			switch moveType {
			case 100:
				// d
				gameField.CurrentPiece.MoveLeft()
			case 97:
				// a
				gameField.CurrentPiece.MoveRight()
			case 115:
				// s
				gameField.CurrentPiece.MoveDown()
			case 113:
				// q
				gameField.CurrentPiece.Rotate(field.Left)
			case 101:
				// e
				gameField.CurrentPiece.Rotate(field.Right)
			}
		case <-timeout:
			return
		}
	}
}

func FormatFieldMessage(isEnemyField int, isAlive int, gameField *field.Field) string {
	return fmt.Sprintf("%d %d %s %d %d %d %d", isEnemyField, isAlive, gameField.String(), gameField.GetSpeed(), *gameField.Score, *gameField.CleanCount, gameField.NextPiece.PieceType)
}
