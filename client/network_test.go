package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	lib "github.com/vszholobov/tetrisLib"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNetworkHTTPVersionErrorsAndBodyClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Client-Version") != "3.0.0" {
			t.Error("missing wire version")
		}
		if r.URL.Path == "/down" {
			w.WriteHeader(503)
			return
		}
		json.NewEncoder(w).Encode([]SessionDto{{SessionId: 42}})
	}))
	defer server.Close()
	var sessions []SessionDto
	if err := networkJSON(server.Client(), server.URL, &sessions); err != nil || len(sessions) != 1 || sessions[0].SessionId != 42 {
		t.Fatalf("sessions: %v %v", sessions, err)
	}
	if err := networkJSON(server.Client(), server.URL+"/down", &sessions); err == nil || !strings.Contains(err.Error(), "503") {
		t.Fatalf("bad HTTP error: %v", err)
	}
	server.Close()
	if err := networkJSON(server.Client(), server.URL, &sessions); err == nil || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("bad offline error: %v", err)
	}
}
func TestNetworkGameSendsCommandAndReceivesResult(t *testing.T) {
	serverErrors := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			serverErrors <- err
			return
		}
		defer c.Close()
		c.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, command, err := c.ReadMessage()
		if err != nil {
			serverErrors <- err
			return
		}
		if len(command) != 1 || command[0] != byte(lib.MoveDown) {
			serverErrors <- errors.New("incorrect command")
			return
		}
		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, lib.GameStateMessage{GameResult: lib.Win, FieldType: lib.Self})
		serverErrors <- c.WriteMessage(websocket.BinaryMessage, buf.Bytes())
	}))
	defer server.Close()
	c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	keys := make(chan rune, 1)
	keys <- 's'
	var output bytes.Buffer
	if err := playNetwork(&output, c, keys, make(chan error), true); err != nil {
		t.Fatal(err)
	}
	if err := <-serverErrors; err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "NETWORK INPUT key=s") || !strings.Contains(output.String(), "Game result: 1") {
		t.Fatalf("missing exchange: %s", output.String())
	}
}
func TestNetworkMenuAndCompactRendering(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]SessionDto{{SessionId: 1, Started: true}, {SessionId: 2}, {SessionId: 3}})
	}))
	defer server.Close()
	keys := make(chan rune, 2)
	keys <- 's'
	keys <- '\r'
	var output bytes.Buffer
	id, create, err := networkMenu(&output, server.Client(), strings.TrimPrefix(server.URL, "http://"), keys, make(chan error))
	if err != nil || create || id != "3" {
		t.Fatalf("selection: %s %t %v", id, create, err)
	}
	var previous [lib.FieldHeight + 2]string
	rows := networkRows(lib.GameStateMessage{}, lib.GameStateMessage{})
	for _, row := range rows {
		if len(row) > 80 {
			t.Fatalf("row exceeds framebuffer terminal: %q", row)
		}
	}
	if err := drawNetworkRows(&output, &previous, rows); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err := drawNetworkRows(&output, &previous, rows); err != nil || output.Len() != 0 {
		t.Fatal("unchanged frame emitted")
	}
	rows[1] += "#"
	failing := &networkShortWriter{}
	saved := previous
	if err := drawNetworkRows(failing, &previous, rows); err != io.ErrShortWrite || saved != previous {
		t.Fatal("failed write updated cache")
	}
}

type networkShortWriter struct{}

func (*networkShortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }
func TestNetworkPacketValidation(t *testing.T) {
	for n := 0; n < 50; n++ {
		_, err := decodeGameStateMessage(make([]byte, n))
		if (n == 40) != (err == nil) {
			t.Fatalf("size %d", n)
		}
	}
}
