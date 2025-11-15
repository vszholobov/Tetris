package lib

type PieceType int

var PieceTypeCount int = 7

const (
	TShape      PieceType = 0
	ZigZagLeft  PieceType = 1
	ZigZagRight PieceType = 2
	RightLShape PieceType = 3
	LeftLShape  PieceType = 4
	IShape      PieceType = 5
	SquareShape PieceType = 6
)

type GameStateMessage struct {
	isEnemyField bool
	isAlive      bool
	fieldVal     string
	speed        int
	score        int
	cleanCount   int
	nextPiece    PieceType
}
