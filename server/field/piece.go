package field

type PieceType int

const (
	TShape      PieceType = 0
	ZigZagLeft  PieceType = 1
	ZigZagRight PieceType = 2
	RightLShape PieceType = 3
	LeftLShape  PieceType = 4
	IShape      PieceType = 5
	SquareShape PieceType = 6
)

var rotationsCntByType = map[PieceType]int{
	TShape:      4,
	ZigZagLeft:  2,
	ZigZagRight: 2,
	RightLShape: 4,
	LeftLShape:  4,
	IShape:      2,
	SquareShape: 1,
}

type RotationType int

const (
	PieceRotateLeft  RotationType = -1
	PieceRotateRight RotationType = 1
)

type PieceMoveDirection int

const (
	PieceMoveLeft  PieceMoveDirection = 0
	PieceMoveRight PieceMoveDirection = 1
	PieceMoveDown  PieceMoveDirection = 2
)
