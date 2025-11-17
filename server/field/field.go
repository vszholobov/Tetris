package field

import (
	"strings"

	lib "github.com/vszholobov/tetrisLib"
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
	GetNextPieceType() lib.PieceType
	String() string
}

var fullLineString = strings.Repeat("1", lib.FieldWidth)
var emptyLineString = "1" + strings.Repeat("0", lib.FieldWidth-2) + "1"
var defaultFieldString = buildDefaultFieldString()

func buildDefaultFieldString() string {
	var defaultFieldString strings.Builder
	defaultFieldString.WriteString(fullLineString)
	for i := 1; i < lib.FieldHeight; i++ {
		defaultFieldString.WriteString(emptyLineString)
	}
	return defaultFieldString.String()
}

func MakeDefaultField(pieceSelector PieceSelector) Field {
	return MakeField(pieceSelector, defaultFieldString)
}

func MakeField(pieceSelector PieceSelector, fieldVal string) Field {
	gameField := makeBigIntField(pieceSelector, fieldVal)
	gameField.SelectNextPiece()
	gameField.SelectNextPiece()
	return gameField
}
