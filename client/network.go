package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/mattn/go-tty"
	lib "github.com/vszholobov/tetrisLib"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"tetrisClient/keyboard"
	"time"
)

// Network play uses terminal bytes for exit, including without Unix signal delivery.
func runNetwork(args []string) error {
	flags := flag.NewFlagSet("--network", flag.ContinueOnError)
	address := flags.String("server", serverAddress, "server host:port (HTTP/WebSocket)")
	create := flags.Bool("create", false, "create a session without the menu")
	session := flags.String("session", "", "join a session without the menu")
	trace := flags.Bool("trace", false, "log integration acknowledgements")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 || *address == "" || (*create && *session != "") {
		return fmt.Errorf("invalid network arguments")
	}
	transport := &http.Transport{Proxy: nil, MaxIdleConns: 2, IdleConnTimeout: 10 * time.Second}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	terminal, err := tty.Open()
	if err != nil {
		return fmt.Errorf("open terminal: %w", err)
	}
	defer terminal.Close()
	if _, err := terminal.Raw(); err != nil {
		return fmt.Errorf("raw terminal: %w", err)
	}
	if err := keyboard.ClearScreen(); err != nil {
		return err
	}
	fmt.Print("\x1b[?25l")
	defer fmt.Print("\x1b[?25h")
	keys := make(chan rune, 64)
	errors := make(chan error, 1)
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			key, err := terminal.ReadRune()
			if err != nil {
				select {
				case errors <- err:
				case <-done:
				}
				return
			}
			select {
			case keys <- key:
			case <-done:
				return
			}
		}
	}()
	id := *session
	if !*create && id == "" {
		var chooseCreate bool
		id, chooseCreate, err = networkMenu(os.Stdout, client, *address, keys, errors)
		if err != nil || (id == "" && !chooseCreate) {
			return err
		}
		*create = chooseCreate
	}
	if *create {
		var response CreateSessionResponse
		if err := networkJSON(client, "http://"+*address+"/session/create", &response); err != nil {
			return err
		}
		id = strconv.FormatInt(response.SessionId, 10)
	}
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		return fmt.Errorf("invalid session ID")
	}
	dialer := websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	connection, _, err := dialer.Dial("ws://"+*address+"/session/connect/"+id, nil)
	if err != nil {
		return fmt.Errorf("connect to Tetris server: %w", err)
	}
	defer connection.Close()
	fmt.Printf("\x1b[1;1HSession %s | waiting for second player\x1b[K", id)
	if *trace {
		fmt.Printf("\r\nNETWORK CONNECTED session=%s\r\n", id)
	}
	return playNetwork(os.Stdout, connection, keys, errors, *trace)
}

func networkJSON(client *http.Client, url string, output any) error {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("X-Client-Version", "3.0.0")
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("Tetris server unavailable: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Tetris server: %s", response.Status)
	}
	return json.NewDecoder(io.LimitReader(response.Body, 1024*1024)).Decode(output)
}
func networkMenu(out io.Writer, client *http.Client, address string, keys <-chan rune, errors <-chan error) (string, bool, error) {
	var sessions []SessionDto
	load := func() error {
		var all []SessionDto
		if err := networkJSON(client, "http://"+address+"/session", &all); err != nil {
			return err
		}
		sessions = nil
		for _, s := range all {
			if !s.Started {
				sessions = append(sessions, s)
			}
		}
		return nil
	}
	if err := load(); err != nil {
		return "", false, err
	}
	selected := 0
	for {
		fmt.Fprint(out, "\x1b[H\x1b[2JTetris network | c: create, r: refresh, w/s: select, Enter: join, x: quit\r\n")
		for index, s := range sessions {
			marker := " "
			if index == selected {
				marker = ">"
			}
			fmt.Fprintf(out, "%s Session %d\r\n", marker, s.SessionId)
		}
		select {
		case err := <-errors:
			return "", false, err
		case key := <-keys:
			switch key {
			case 'x', 'e', 27, 3:
				return "", false, nil
			case 'c':
				return "", true, nil
			case 'r':
				if err := load(); err != nil {
					return "", false, err
				}
				selected = 0
			case 'w':
				if selected > 0 {
					selected--
				}
			case 's':
				if selected+1 < len(sessions) {
					selected++
				}
			case '\r', '\n':
				if len(sessions) != 0 {
					return strconv.FormatInt(sessions[selected].SessionId, 10), false, nil
				}
			}
		}
	}
}

func playNetwork(out io.Writer, connection *websocket.Conn, keys <-chan rune, inputErrors <-chan error, trace bool) error {
	states := make(chan lib.GameStateMessage, 64)
	failures := make(chan error, 1)
	done := make(chan struct{})
	defer close(done)
	pongs := make(chan string, 1)
	connection.SetPongHandler(func(data string) error {
		select {
		case pongs <- data:
		default:
		}
		return nil
	})
	go func() {
		for {
			_, data, err := connection.ReadMessage()
			if err != nil {
				select {
				case failures <- err:
				case <-done:
				}
				return
			}
			state, err := decodeGameStateMessage(data)
			if err != nil {
				select {
				case failures <- err:
				case <-done:
				}
				return
			}
			if state.FieldType != lib.Self && state.FieldType != lib.Enemy {
				select {
				case failures <- fmt.Errorf("invalid field type"):
				case <-done:
				}
				return
			}
			select {
			case states <- *state:
			case <-done:
				return
			}
		}
	}()
	var own, enemy lib.GameStateMessage
	var previous [lib.FieldHeight + 2]string
	dirty := false
	frames := time.NewTicker(time.Second / 30)
	defer frames.Stop()
	ping := time.NewTicker(3 * time.Second)
	defer ping.Stop()
	for {
		select {
		case data := <-pongs:
			if trace {
				fmt.Fprintf(out, "\r\nNETWORK PONG %s\r\n", data)
			}
		case err := <-failures:
			return fmt.Errorf("Tetris connection closed: %w", err)
		case err := <-inputErrors:
			return err
		case state := <-states:
			if state.FieldType == lib.Self {
				own = state
			} else {
				enemy = state
			}
			dirty = true
			if trace {
				fmt.Fprintf(out, "\r\nNETWORK FIELD type=%d data=%x\r\n", state.FieldType, state.FieldBytes)
			}
			if state.GameResult != lib.Ongoing {
				fmt.Fprintf(out, "\x1b[%d;1HGame result: %d\r\n", lib.FieldHeight+3, state.GameResult)
				return nil
			}
		case key := <-keys:
			if key == 'x' || key == 27 || key == 3 {
				_ = connection.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
				fmt.Fprintf(out, "\x1b[%d;1H\r\nNETWORK EXIT\r\n", lib.FieldHeight+3)
				return nil
			}
			command, ok := map[rune]lib.ClientCommand{'a': lib.MoveLeft, 'd': lib.MoveRight, 's': lib.MoveDown, 'q': lib.RotateLeft, 'e': lib.RotateRight}[key]
			if ok {
				if err := connection.WriteMessage(websocket.TextMessage, []byte{byte(command)}); err != nil {
					return err
				}
				if trace {
					fmt.Fprintf(out, "\r\nNETWORK INPUT key=%c\r\n", key)
				}
			}
		case <-ping.C:
			if err := connection.WriteControl(websocket.PingMessage, []byte("tetris"), time.Now().Add(time.Second)); err != nil {
				return err
			}
		case <-frames.C:
			if dirty {
				rows := networkRows(own, enemy)
				if err := drawNetworkRows(out, &previous, rows); err != nil {
					return err
				}
				dirty = false
			}
		}
	}
}
func networkRows(own, enemy lib.GameStateMessage) [lib.FieldHeight + 2]string {
	left, right := singleRows(own), singleRows(enemy)
	left[0] = "Tetris | a/d/s/q/e: play, x: quit"
	for i := 1; i <= lib.FieldHeight; i++ {
		left[i] = fmt.Sprintf("%-32s  %s", left[i], right[i])
	}
	left[len(left)-1] = fmt.Sprintf("Self: %d | Enemy: %d | Speed: %d | Lines: %d", own.Score, enemy.Score, own.Speed, own.CleanCount)
	return left
}
func drawNetworkRows(out io.Writer, previous *[lib.FieldHeight + 2]string, rows [lib.FieldHeight + 2]string) error {
	var frame strings.Builder
	for row, next := range rows {
		old := previous[row]
		if next == old {
			continue
		}
		width := max(len(old), len(next))
		fmt.Fprintf(&frame, "\x1b[%d;1H%s%s", row+1, next, strings.Repeat(" ", width-len(next)))
	}
	if frame.Len() != 0 {
		n, err := io.WriteString(out, frame.String())
		if err != nil {
			return err
		}
		if n != frame.Len() {
			return io.ErrShortWrite
		}
	}
	*previous = rows
	return nil
}
