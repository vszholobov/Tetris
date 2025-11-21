package main

import (
	"math/big"
	"strconv"
	"sync"
	"tetrisClient/keyboard"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	lib "github.com/vszholobov/tetrisLib"
)

type Session struct {
	conn             *websocket.Conn
	inputProcessor   *keyboard.InputProcessor
	outputController *keyboard.OutputController
	endSessionMutex  sync.Mutex
	sendMessageMutex sync.Mutex
	isSessionEnded   bool
	pingMeasurer     *PingMeasurer
}

func MakeGameSession(connect *websocket.Conn, inputProcessor *keyboard.InputProcessor, outputController *keyboard.OutputController) *Session {
	gameSession = &Session{
		inputProcessor:   inputProcessor,
		outputController: outputController,
		pingMeasurer:     MakePingMeasurer(),
		conn:             connect,
	}
	connect.SetPongHandler(gameSession.pingMeasurer.pongHandler())
	connect.SetCloseHandler(func(code int, text string) error {
		onExit(strconv.Itoa(code), outputController)
		return nil
	})
	return gameSession
}

func (gameSession *Session) processPlayerPing() {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()
	for range ticker.C {
		pingUuid := gameSession.pingMeasurer.addMeasure()
		if err := gameSession.sendPingMessage(pingUuid); err != nil {
			break
		}
	}
}

func (gameSession *Session) processServerMessages() {
	for {
		_, message, err := gameSession.conn.ReadMessage()
		if err != nil {
			onExit("Connection closed(", gameSession.outputController)
		}
		gameState, err := decodeGameStateMessage(message)
		if err != nil {
			onExit(err.Error(), gameSession.outputController)
		}

		fieldBig := new(big.Int).SetBytes(gameState.FieldBytes[:])
		if gameState.GameResult != lib.Ongoing {
			var exitMessage string
			switch gameState.GameResult {
			case lib.Win:
				exitMessage = "WIN!"
			case lib.Lose:
				exitMessage = "LOSE("
			case lib.Draw:
				exitMessage = "DRAW"
			}
			onExit(exitMessage, gameSession.outputController)
		}
		if gameState.FieldType == lib.Self {
			PrintSelfField(
				fieldBig,
				strconv.Itoa(int(gameState.Speed)),
				strconv.Itoa(int(gameState.Score)),
				strconv.Itoa(int(gameState.CleanCount)),
				gameState.NextPiece,
				gameSession.pingMeasurer.actualPing.String(),
			)
		} else {
			PrintEnemyField(
				fieldBig,
				strconv.Itoa(int(gameState.Speed)),
				strconv.Itoa(int(gameState.Score)),
				strconv.Itoa(int(gameState.CleanCount)),
				gameState.NextPiece,
			)
		}
	}
}

func (gameSession *Session) processPlayerActions() {
	for {
		messageFromKeyboard := gameSession.inputProcessor.Read()
		var playerCommand lib.ClientCommand
		switch messageFromKeyboard {
		case 'a':
			playerCommand = lib.MoveLeft
		case 's':
			playerCommand = lib.MoveDown
		case 'd':
			playerCommand = lib.MoveRight
		case 'q':
			playerCommand = lib.RotateLeft
		case 'e':
			playerCommand = lib.RotateRight
		default:
			continue
		}
		err := gameSession.sendMessage([]byte{uint8(playerCommand)})
		if err != nil {
			return
		}
	}
}

func (gameSession *Session) sendPingMessage(pingId uuid.UUID) error {
	gameSession.sendMessageMutex.Lock()
	defer gameSession.sendMessageMutex.Unlock()
	return gameSession.conn.WriteMessage(websocket.PingMessage, pingId[:])
}

func (gameSession *Session) sendMessage(message []byte) error {
	gameSession.sendMessageMutex.Lock()
	defer gameSession.sendMessageMutex.Unlock()
	return gameSession.conn.WriteMessage(websocket.TextMessage, message)
}
