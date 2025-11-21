package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"tetrisClient/keyboard"

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
`

func main() {
	outputController := keyboard.InitOutputController()
	handleSigtermExit(outputController)
	if len(os.Args) >= 2 {
		onExit(helpMessage, outputController)
	}

	outputController.HideCursor()
	defer outputController.ShowCursor()

	inputProcessor := keyboard.MakeInputProcessor()
	defer inputProcessor.Close()

	go inputProcessor.ProcessKeyboardInput()

	var sessionId string
	menu := MakeMenu(inputProcessor, outputController)
	menu.handleMenu()
	if menu.isExit {
		onExit("", outputController)
	} else if menu.isCreateSession {
		sessionId = createSession()
	} else {
		sessionId = menu.getSelectedSessionId()
	}
	sessionConnectUrl := url.URL{Scheme: "ws", Host: serverAddress, Path: "/session/connect/" + sessionId}

	connect, _, _ := websocket.DefaultDialer.Dial(sessionConnectUrl.String(), nil)
	defer connect.Close()
	gameSession := MakeGameSession(connect, inputProcessor, outputController)
	fmt.Println("SessionId: " + sessionId)

	go gameSession.processServerMessages()
	go gameSession.processPlayerPing()
	gameSession.processPlayerActions()
}

func createSession() string {
	response, err := http.Get("http://" + serverAddress + "/session/create")
	if err != nil {
		panic(err.Error())
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err.Error())
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
	body, readResponseError := io.ReadAll(response.Body)
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

// onExit Closes keyboard input stream and makes cursor visible back
func onExit(exitMessage string, outputController *keyboard.OutputController) {
	gameSession.endSessionMutex.Lock()
	if !gameSession.isSessionEnded {
		gameSession.isSessionEnded = true
		outputController.ShowCursor()
		outputController.Clear()
		fmt.Println(exitMessage)
		if gameSession.inputProcessor != nil {
			gameSession.inputProcessor.Close()
		}
		if gameSession.conn != nil {
			gameSession.conn.Close()
		}
	}
	gameSession.endSessionMutex.Unlock()
	os.Exit(0)
}
