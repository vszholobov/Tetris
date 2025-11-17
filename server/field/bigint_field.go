package field

import (
	"math/big"
	"tetrisServer/env"

	lib "github.com/vszholobov/tetrisLib"
)

var fullLine, _ = big.NewInt(0).SetString(fullLineString, 2)
var emptyLine, _ = big.NewInt(0).SetString(emptyLineString, 2)

type BigIntField struct {
	Val           *big.Int
	CurrentPiece  *bigIntPiece
	NextPiece     *bigIntPiece
	Score         *int
	CleanCount    *int
	pieceSelector PieceSelector
}

func makeBigIntField(pieceSelector PieceSelector, fieldString string) *BigIntField {
	fieldVal, _ := big.NewInt(0).SetString(fieldString, 2)
	return &BigIntField{
		Val:           fieldVal,
		Score:         new(int),
		CleanCount:    new(int),
		pieceSelector: pieceSelector,
	}
}

func (gameField *BigIntField) SelectNextPiece() {
	pieceType := gameField.pieceSelector.SelectNextPiece()
	piece := makePiece(gameField, pieceType)
	gameField.CurrentPiece = gameField.NextPiece
	gameField.NextPiece = &piece
}

func (gameField *BigIntField) CleanLines() {
	restField := big.NewInt(0)
	currentCleanCount := 0
	for i := 0; i < lib.FieldHeight-1; i++ {
		curRange := uint(i * lib.FieldWidth)
		lineMask := big.NewInt(0).Lsh(fullLine, curRange)
		lineIsFilled := big.NewInt(0).And(lineMask, gameField.Val).Cmp(lineMask) == 0

		if lineIsFilled {
			// add empy line to end of field
			restField.Lsh(restField, lib.FieldWidth)
			restField.Or(restField, emptyLine)
			currentCleanCount += 1
		} else {
			// add current line to start of field
			lineMask.And(lineMask, gameField.Val)
			restField.Or(lineMask, restField)
		}
	}
	*gameField.CleanCount += currentCleanCount
	*gameField.Score += currentCleanCount * gameField.GetSpeed() * 10 / (5 - currentCleanCount)
	// 22 lines. One redundant line for correct or concatenation.
	// So shift to the right by the length of the field after concatenation to remove redundant empty line
	gameField.Val.SetString(
		"111111111111"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000"+
			"000000000000", 2)
	gameField.Val.Or(gameField.Val, restField)
}

func (gameField *BigIntField) ConcatPiece() {
	gameField.Val.Or(gameField.Val, gameField.CurrentPiece.getVal())
}

func (gameField *BigIntField) GetSpeed() int {
	return *gameField.CleanCount/env.CleanRowsCountToIncreaseSpeed + 1
}

func (gameField *BigIntField) GetScore() int {
	return *gameField.Score
}

func (gameField *BigIntField) GetCleanCount() int {
	return *gameField.CleanCount
}

func (gameField *BigIntField) GetNextPieceType() lib.PieceType {
	return gameField.NextPiece.PieceType
}

func (gameField *BigIntField) MovePiece(moveDirection PieceMoveDirection) bool {
	return gameField.CurrentPiece.move(moveDirection)
}

func (gameField *BigIntField) CanMovePiece(moveDirection PieceMoveDirection) bool {
	return gameField.CurrentPiece.canMove(moveDirection)
}

func (gameField *BigIntField) RotatePiece(rotationType RotationType) bool {
	return gameField.CurrentPiece.rotate(rotationType)
}

func (gameField *BigIntField) String() string {
	newField := big.NewInt(0).Or(gameField.Val, gameField.CurrentPiece.getVal())
	return newField.String()
}

func (gameField *BigIntField) intersects(pieceVal *big.Int) bool {
	newField := big.NewInt(0).Set(gameField.Val)
	newShape := big.NewInt(0).Set(pieceVal)
	return newField.And(newField, newShape).Cmp(big.NewInt(0)) != 0
}
