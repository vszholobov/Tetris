package field

import (
	"strings"

	lib "github.com/vszholobov/tetrisLib"
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
	GetNextPieceType() lib.PieceType
	String() string
}

var defaultFieldString = strings.ReplaceAll(`
111111111111
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
100000000001
`, "\n", "")

func MakeDefaultField(pieceSelector PieceSelector) Field {
	return MakeField(pieceSelector, defaultFieldString)
}

func MakeField(pieceSelector PieceSelector, fieldVal string) Field {
	gameField := makeBigIntField(pieceSelector, fieldVal)
	gameField.SelectNextPiece()
	gameField.SelectNextPiece()
	return gameField
}
