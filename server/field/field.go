package field

import (
	"math/big"
	"math/rand"
)

const FieldWidth = 12
const FieldHeight = 21
const CleanRowsCountToIncreaseSpeed = 40

var fullLine, _ = big.NewInt(0).SetString("111111111111", 2)
var emptyLine, _ = big.NewInt(0).SetString("100000000001", 2)

type Field struct {
	Val            *big.Int
	CurrentPiece   *Piece
	NextPiece      *Piece
	Score          *int
	CleanCount     *int
	pieceGenerator *rand.Rand
}

func MakeDefaultField(pieceGenerator *rand.Rand) Field {
	fieldVal, _ := big.NewInt(0).SetString(
		"111111111111"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001"+
			"100000000001", 2)
	gameField := MakeField(fieldVal, pieceGenerator)
	gameField.SelectNextPiece()
	gameField.SelectNextPiece()
	return gameField
}

func MakeField(fieldVal *big.Int, pieceGenerator *rand.Rand) Field {
	score := 0
	cleanCount := 0
	return Field{Val: fieldVal, Score: &score, CleanCount: &cleanCount, pieceGenerator: pieceGenerator}
}

func (gameField *Field) SelectNextPiece() {
	pieceTypeRnd := gameField.pieceGenerator.Intn(7)
	var pieceType PieceType
	if pieceTypeRnd == 0 {
		pieceType = IShape
	} else if pieceTypeRnd == 1 {
		pieceType = RightLShape
	} else if pieceTypeRnd == 2 {
		pieceType = TShape
	} else if pieceTypeRnd == 3 {
		pieceType = ZigZagRight
	} else if pieceTypeRnd == 4 {
		pieceType = ZigZagLeft
	} else if pieceTypeRnd == 5 {
		pieceType = SquareShape
	} else if pieceTypeRnd == 6 {
		pieceType = LeftLShape
	}
	piece := MakePiece(gameField, pieceType)
	gameField.CurrentPiece = gameField.NextPiece
	gameField.NextPiece = &piece
}

func (gameField *Field) String() string {
	newField := big.NewInt(0).Or(gameField.Val, gameField.CurrentPiece.GetVal())
	return newField.String()
}

func (gameField *Field) Clean() {
	restField := big.NewInt(0)
	currentCleanCount := 0
	for i := 0; i < FieldHeight-1; i++ {
		curRange := uint(i * FieldWidth)
		lineMask := big.NewInt(0).Lsh(fullLine, curRange)
		lineIsFilled := big.NewInt(0).And(lineMask, gameField.Val).Cmp(lineMask) == 0

		if lineIsFilled {
			// add empy line to end of field
			restField.Lsh(restField, FieldWidth)
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

func (gameField *Field) Intersects(pieceVal *big.Int) bool {
	newField := big.NewInt(0).Set(gameField.Val)
	newShape := big.NewInt(0).Set(pieceVal)
	return newField.And(newField, newShape).Cmp(big.NewInt(0)) != 0
}

func (gameField *Field) GetSpeed() int {
	return *gameField.CleanCount/CleanRowsCountToIncreaseSpeed + 1
}
