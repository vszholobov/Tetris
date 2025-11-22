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

var Version = "dev"

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
	keyboard := keyboard.MakeKeyboard()
	defer keyboard.Close()
	gameSession := MakeGameSession(keyboard)
	handleSigtermExit()
	if len(os.Args) >= 2 {
		onExit(helpMessage)
	}

	keyboard.HideCursor()
	defer keyboard.ShowCursor()

	go keyboard.ProcessKeyboardInput()

	var sessionId string
	menu := MakeMenu(keyboard)
	menu.handleMenu()
	if menu.isExit {
		onExit("")
	} else if menu.isCreateSession {
		sessionId = createSession()
	} else {
		sessionId = menu.getSelectedSessionId()
	}
	sessionConnectUrl := url.URL{Scheme: "ws", Host: serverAddress, Path: "/session/connect/" + sessionId}

	connect, _, err := websocket.DefaultDialer.Dial(sessionConnectUrl.String(), nil)
	if err != nil {
		onExit("Connection error(")
	}
	defer connect.Close()
	gameSession.SetConnection(connect)
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
	req, err := http.NewRequest("GET", "http://"+serverAddress+"/session", nil)
	if err != nil {
		onExit(err.Error())
	}

	req.Header.Set("X-Client-Version", Version)

	resp, err := http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		onExit(resp.Status)
	}
	if err != nil {
		onExit(err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		onExit(err.Error())
	}

	var listSessions []SessionDto
	err = json.Unmarshal(body, &listSessions)
	if err != nil {
		onExit(err.Error())
	}

	activeSessions := make([]SessionDto, 0)
	for _, s := range listSessions {
		if !s.Started {
			activeSessions = append(activeSessions, s)
		}
	}
	return activeSessions
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
func handleSigtermExit() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		onExit("")
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
func onExit(exitMessage string) {
	if gameSession.Close() {
		fmt.Println(exitMessage)
		os.Exit(0)
	}
}
