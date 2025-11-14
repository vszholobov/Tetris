package field

import (
	"math/rand"
)

const FieldWidth = 12
const FieldHeight = 21
const CleanRowsCountToIncreaseSpeed = 12

type Field interface {
	SelectNextPiece()
	CleanLines()
	ConcatPiece()
	MovePiece(moveDirection PieceMoveDirection) bool
	CanMovePiece(moveDirection PieceMoveDirection) bool
	RotatePiece(rotationType RotationType) bool
	GetSpeed() int
	GetScore() int
	GetCleanCount() int
	GetNextPieceType() PieceType
	String() string
}

func MakeField(pieceGenerator *rand.Rand) Field {
	gameField := makeBigIntField(pieceGenerator)
	gameField.SelectNextPiece()
	gameField.SelectNextPiece()
	return gameField
}
