package server

import (
	"math/rand"
	"testing"
	"tetrisServer/field"
	"time"

	"github.com/gorilla/websocket"
)

type MockConn struct {
	Written [][]byte
	Closed  bool
}

func (m *MockConn) WriteMessage(mt int, data []byte) error {
	m.Written = append(m.Written, data)
	return nil
}

func (m *MockConn) ReadMessage() (int, []byte, error) {
	return websocket.TextMessage, []byte("d"), nil
}

func (m *MockConn) SetWriteDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error   { return nil }
func (m *MockConn) Close() error                        { m.Closed = true; return nil }
func (m *MockConn) SetPongHandler(f func(string) error) {}

func TestRaceEndSession(t *testing.T) {
	gameField1 := field.MakeDefaultField(field.MakePieceSelector(rand.New(rand.NewSource(1))))
	gameField2 := field.MakeDefaultField(field.MakePieceSelector(rand.New(rand.NewSource(1))))

	for i := 0; i < 10000; i++ {
		ps1 := &PlayerSession{playerField: gameField1, conn: &MockConn{}, gameSession: &GameSession{}}
		ps2 := &PlayerSession{playerField: gameField2, conn: &MockConn{}, gameSession: &GameSession{}}

		ps1.enemySession = ps2
		ps2.enemySession = ps1

		done := make(chan struct{}, 2)

		go func() {
			ps1.endSession()
			done <- struct{}{}
		}()
		go func() {
			ps2.endSession()
			done <- struct{}{}
		}()

		<-done
		<-done
	}
}
