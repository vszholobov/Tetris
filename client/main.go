package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"tetrisClient/keyboard"
	"time"

	"github.com/gorilla/websocket"
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
	pingMs                 uint64
	endSessionMutex        sync.Mutex
	isSessionEnded         bool
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
	gameSession = &Session{keyboardInputProcessor: inputProcessor}

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
	connect.SetPingHandler(func(appData string) error {
		return connect.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second*10))
	})
	connect.SetCloseHandler(func(code int, text string) error {
		onExit(strconv.Itoa(code), outputController)
		return nil
	})
	gameSession.conn = connect
	defer gameSession.conn.Close()
	fmt.Println("SessionId: " + sessionId)

	go readProcessor(connect, outputController)
	sendProcessor(connect, inputProcessor.GetKeyboardInputTransferChannel())
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
	c *websocket.Conn,
	keyboardSendChannel chan rune,
) {
	for messageFromKeyboard := range keyboardSendChannel {
		err := c.WriteMessage(websocket.TextMessage, []byte(string(messageFromKeyboard)))
		if err != nil {
			// log.Println("write:", err)
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

// readProcessor server handler
func readProcessor(c *websocket.Conn, outputController *keyboard.OutputController) {
	func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				onExit("Connection closed(", outputController)
			}
			fields := strings.Fields(string(message))
			if fields[0] == "0" {
				// self field
				if fields[1] == "0" {
					onExit(fields[2], outputController)
				}
				field, _ := big.NewInt(0).SetString(string(fields[2]), 10)
				speed := fields[3]
				score := fields[4]
				cleanCount := fields[5]
				nextPieceTypeIntRepr, _ := strconv.Atoi(fields[6])
				nextPieceType := PieceType(nextPieceTypeIntRepr)
				PrintSelfField(field, speed, score, cleanCount, nextPieceType, getPingRepresentation())
			} else if fields[0] == "1" {
				// enemy field
				if fields[1] == "0" {

				}
				field, _ := big.NewInt(0).SetString(string(fields[2]), 10)
				speed := fields[3]
				score := fields[4]
				cleanCount := fields[5]
				nextPieceTypeIntRepr, _ := strconv.Atoi(fields[6])
				nextPieceType := PieceType(nextPieceTypeIntRepr)
				PrintEnemyField(field, speed, score, cleanCount, nextPieceType)
			} else {
				gameSession.pingMs, _ = strconv.ParseUint(fields[1], 10, 64)
			}
		}
	}()
}

func getPingRepresentation() string {
	if gameSession.pingMs < 1000 {
		return strconv.FormatUint(gameSession.pingMs, 10) + "ms"
	} else {
		return fmt.Sprintf("%.1fs", float64(gameSession.pingMs)/1000)
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
