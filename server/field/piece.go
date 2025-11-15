package field

import "math/rand"

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

var PieceMoveDirectionCount int = 3

const (
	PieceMoveLeft  PieceMoveDirection = 0
	PieceMoveRight PieceMoveDirection = 1
	PieceMoveDown  PieceMoveDirection = 2
)

type PieceSelector interface {
	SelectNextPiece() PieceType
}

type DefaultPieceSelector struct {
	random *rand.Rand
}

func MakePieceSelector(random *rand.Rand) PieceSelector {
	return &DefaultPieceSelector{random: random}
}

func (ps *DefaultPieceSelector) SelectNextPiece() PieceType {
	random := ps.random.Intn(PieceTypeCount)
	switch random {
	case 0:
		return IShape
	case 1:
		return RightLShape
	case 2:
		return TShape
	case 3:
		return ZigZagRight
	case 4:
		return ZigZagLeft
	case 5:
		return SquareShape
	case 6:
		return LeftLShape
	default:
		panic("Unknown piece type")
	}
}
