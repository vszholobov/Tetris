package lib

import (
	"fmt"
	"strconv"
)

const FieldWidth = 12
const FieldHeight = 21

type Ping uint16

func (ping Ping) String() string {
	if ping < 1000 {
		return strconv.FormatUint(uint64(ping), 10) + "ms"
	} else {
		return fmt.Sprintf("%.1fs", float64(ping)/1000)
	}
}

var PieceTypeCount uint8 = 7

type PieceType uint8

const (
	TShape      PieceType = 0
	ZigZagLeft  PieceType = 1
	ZigZagRight PieceType = 2
	RightLShape PieceType = 3
	LeftLShape  PieceType = 4
	IShape      PieceType = 5
	SquareShape PieceType = 6
)

type GameResult uint8

const (
	Ongoing GameResult = 0
	Win     GameResult = 1
	Lose    GameResult = 2
	Draw    GameResult = 3
)

type FieldType uint8

const (
	Self  FieldType = 0
	Enemy FieldType = 1
)

type FieldBytes [32]byte

// size = 40 byte
type GameStateMessage struct {
	FieldType  FieldType
	GameResult GameResult
	FieldBytes FieldBytes
	Speed      uint8
	Score      uint16
	CleanCount uint16
	NextPiece  PieceType
}
