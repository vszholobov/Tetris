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
	GetSpeed() uint8
	GetScore() uint16
	GetCleanCount() uint16
	GetNextPieceType() lib.PieceType
	Bytes() lib.FieldBytes
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

// MakeDefaultFieldWithSpeedRows preserves the server's configurable speed progression.
func MakeDefaultFieldWithSpeedRows(pieceSelector PieceSelector, rows uint8) Field {
	f := MakeDefaultField(pieceSelector).(*BigIntField)
	if rows != 0 {
		f.cleanRowsForSpeed = uint16(rows)
	}
	return f
}

func MakeField(pieceSelector PieceSelector, fieldVal string) Field {
	gameField := makeBigIntField(pieceSelector, fieldVal)
	gameField.SelectNextPiece()
	gameField.SelectNextPiece()
	return gameField
}
