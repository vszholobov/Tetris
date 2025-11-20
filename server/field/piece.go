package field

import (
	"math/rand"

	lib "github.com/vszholobov/tetrisLib"
)

var rotationsCntByType = map[lib.PieceType]int{
	lib.TShape:      4,
	lib.ZigZagLeft:  2,
	lib.ZigZagRight: 2,
	lib.RightLShape: 4,
	lib.LeftLShape:  4,
	lib.IShape:      2,
	lib.SquareShape: 1,
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
	SelectNextPiece() lib.PieceType
}

type DefaultPieceSelector struct {
	random *rand.Rand
}

func MakePieceSelector(random *rand.Rand) PieceSelector {
	return &DefaultPieceSelector{random: random}
}

func (ps *DefaultPieceSelector) SelectNextPiece() lib.PieceType {
	random := ps.random.Intn(int(lib.PieceTypeCount))
	return lib.PieceType(random)
}
