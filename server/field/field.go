package field

import (
	"math/big"
	"math/rand"
)

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
	intersects(pieceVal *big.Int) bool
}

func MakeField(pieceGenerator *rand.Rand) Field {
	gameField := makeBigIntField(pieceGenerator)
	gameField.SelectNextPiece()
	gameField.SelectNextPiece()
	return gameField
}
