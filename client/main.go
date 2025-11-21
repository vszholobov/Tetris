package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"tetrisClient/keyboard"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	lib "github.com/vszholobov/tetrisLib"
)

type CreateSessionResponse struct {
	SessionId int64 `json:"sessionId"`
}

type SessionDto struct {
	SessionId int64 `json:"sessionId"`
	Started   bool  `json:"started"`
}

type Session struct {
	conn                   *websocket.Conn
	keyboardInputProcessor *keyboard.InputProcessor
	endSessionMutex        sync.Mutex
	sendMessageMutex       sync.Mutex
	isSessionEnded         bool
	pingMeasurer           *PingMeasurer
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

var serverAddress = "tetris.vszholobov.ru:8080"
var gameSession *Session

var helpMessage = `
Run client menu by launching the application without arguments
List of control keys:
1) Menu
* r - reload running sessions list
* c - create new session
* w - move session list cursor up
* s - move session list cursor down
* enter - connect to selected session
* e - exit game
2) Game
* a - move piece left
* d - move piece right
* s - move piece down
* q - rotate piece left
* e - rotate piece right

It is also available to run the client with command line arguments
* connect <sessionId> - connect to existing session
* create              - create new session
* list                - show list of existing sessions
`

func main() {
	outputController := keyboard.InitOutputController()
	handleSigtermExit(outputController)

	outputController.HideCursor()
	defer outputController.ShowCursor()

	inputProcessor := keyboard.MakeInputProcessor()
	defer inputProcessor.Close()

	go inputProcessor.ProcessKeyboardInput()
	gameSession = &Session{keyboardInputProcessor: inputProcessor, pingMeasurer: MakePingMeasurer()}

	var sessionId string
	if len(os.Args) < 2 {
		menu := MakeMenu(outputController)
		menu.showMenu()
		menu.handleMenu(inputProcessor.GetKeyboardInputTransferChannel())
		if menu.isExit {
			onExit("", outputController)
		}
		if menu.isCreateSession {
			sessionId = createSession()
		} else {
			sessionId = strconv.FormatInt(menu.sessionsList[menu.currentSessionIndex].SessionId, 10)
		}
	} else if operation := os.Args[1]; operation == "connect" {
		sessionId = os.Args[2]
	} else if operation == "create" {
		sessionId = createSession()
	} else if operation == "list" {
		listSessions := getSessionsList()
		fmt.Println("Sessions:")
		for _, session := range listSessions {
			fmt.Printf("Id: %d Started: %t", session.SessionId, session.Started)
			fmt.Println()
		}
		return
	} else if operation == "help" {
		onExit(helpMessage, outputController)
		return
	} else {
		onExit("Operation '"+operation+"' does not exist. See full list by running 'help' operation", outputController)
		return
	}
	sessionConnectUrl := url.URL{Scheme: "ws", Host: serverAddress, Path: "/session/connect/" + sessionId}

	connect, _, _ := websocket.DefaultDialer.Dial(sessionConnectUrl.String(), nil)
	connect.SetPongHandler(gameSession.pingMeasurer.pongHandler())
	connect.SetCloseHandler(func(code int, text string) error {
		onExit(strconv.Itoa(code), outputController)
		return nil
	})
	gameSession.conn = connect
	defer gameSession.conn.Close()
	fmt.Println("SessionId: " + sessionId)

	go readProcessor(connect, outputController)
	go gameSession.processPlayerPing()
	sendProcessor(gameSession, inputProcessor.GetKeyboardInputTransferChannel())
}

func createSession() string {
	response, createSessionError := http.Get("http://" + serverAddress + "/session/create")
	if createSessionError != nil {
		panic(createSessionError.Error())
	}
	body, readResponseError := ioutil.ReadAll(response.Body)

	if readResponseError != nil {
		panic(readResponseError.Error())
	}

	var createSessionResponse CreateSessionResponse
	json.Unmarshal(body, &createSessionResponse)
	return strconv.FormatInt(createSessionResponse.SessionId, 10)
}

func getSessionsList() []SessionDto {
	response, getSessionsListError := http.Get("http://" + serverAddress + "/session")
	if getSessionsListError != nil {
		panic(getSessionsListError.Error())
	}
	body, readResponseError := ioutil.ReadAll(response.Body)
	if readResponseError != nil {
		panic(readResponseError.Error())
	}

	listSessions := make([]SessionDto, 0)
	json.Unmarshal(body, &listSessions)
	return listSessions
}

func sendProcessor(
	gameSession *Session,
	keyboardSendChannel chan rune,
) {
	for messageFromKeyboard := range keyboardSendChannel {
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

// Handles interrupt signal
func handleSigtermExit(outputController *keyboard.OutputController) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		onExit("", outputController)
	}()
}

func decodeGameStateMessage(data []byte) (*lib.GameStateMessage, error) {
	if len(data) != 40 {
		return nil, fmt.Errorf("invalid GameStateMessage size: %d", len(data))
	}
	packet := &lib.GameStateMessage{}
	buf := bytes.NewReader(data)
	if err := binary.Read(buf, binary.BigEndian, packet); err != nil {
		return nil, err
	}
	return packet, nil
}

func readProcessor(c *websocket.Conn, outputController *keyboard.OutputController) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			onExit("Connection closed(", outputController)
		}
		gameState, err := decodeGameStateMessage(message)
		if err != nil {
			onExit(err.Error(), outputController)
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
			onExit(exitMessage, outputController)
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

// onExit Closes keyboard input stream and makes cursor visible back
func onExit(exitMessage string, outputController *keyboard.OutputController) {
	gameSession.endSessionMutex.Lock()
	if !gameSession.isSessionEnded {
		gameSession.isSessionEnded = true
		outputController.ShowCursor()
		outputController.Clear()
		fmt.Println(exitMessage)
		if gameSession.keyboardInputProcessor != nil {
			gameSession.keyboardInputProcessor.Close()
		}
		if gameSession.conn != nil {
			gameSession.conn.Close()
		}
	}
	gameSession.endSessionMutex.Unlock()
	os.Exit(0)
}

type Menu struct {
	currentSessionIndex int
	sessionsList        []SessionDto
	isEnded             bool
	isCreateSession     bool
	isExit              bool
	outputController    *keyboard.OutputController
}

func MakeMenu(outputController *keyboard.OutputController) Menu {
	sessionsList := getSessionsList()
	return Menu{
		currentSessionIndex: 0,
		sessionsList:        sessionsList,
		isEnded:             false,
		isCreateSession:     false,
		isExit:              false,
		outputController:    outputController,
	}
}

func (menu *Menu) showMenu() {
	menu.outputController.Clear()
	fmt.Println(" Tetris🕹️")
	fmt.Println("----------")
	for index, session := range menu.sessionsList {
		currentItem := ""
		if index == menu.currentSessionIndex {
			currentItem += "\033[30;5;107m"
		}
		currentItem += strconv.FormatInt(session.SessionId, 10)
		currentItem += " "
		currentItem += strconv.FormatBool(session.Started)
		if index == menu.currentSessionIndex {
			currentItem += "\033[0m"
		}
		fmt.Println(currentItem)
	}
}

func (menu *Menu) handleMenu(keyboardInputChannel chan rune) {
	for !menu.isEnded {
		input := <-keyboardInputChannel
		switch input {
		case 115:
			// s
			if len(menu.sessionsList) == 0 {
				continue
			}
			menu.currentSessionIndex++
			menu.currentSessionIndex = menu.currentSessionIndex % len(menu.sessionsList)
		case 119:
			// w
			if len(menu.sessionsList) == 0 {
				continue
			}
			menu.currentSessionIndex--
			if menu.currentSessionIndex < 0 {
				menu.currentSessionIndex = len(menu.sessionsList) - 1
			}
		case 114:
			// r
			menu.sessionsList = getSessionsList()
		case 99:
			// c
			menu.isEnded = true
			menu.isCreateSession = true
			continue
		case 13:
			// enter
			if len(menu.sessionsList) == 0 {
				continue
			}
			menu.isEnded = true
			continue
		case 101:
			// e
			menu.isEnded = true
			menu.isExit = true
		default:
			// skip unknown input
			continue
		}
		menu.showMenu()
	}
}
