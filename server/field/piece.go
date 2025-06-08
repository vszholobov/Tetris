package field

import (
	"math"
	"math/big"
)

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

// []
//000001100000
//000001100000 = 393312

// ⅃
//000000100000
//000011100000 = 131296

//000001100000
//000000100000
//000000100000 = 1610743840

//000011100000
//000010000000 = 917632

//000001000000
//000001000000
//000001100000 = 1074004064

// L
//000010000000
//000011100000 = 524512

//000001000000
//000001000000
//000011000000 = 1074004160

//000011100000
//000000100000 = 917536

//000011000000
//000010000000
//000010000000 = 3221749888

// ----
//000011110000 = 240

//000001000000
//000001000000
//000001000000
//000001000000 = 4399120515136

//  --
// --
//000000110000
//000001100000 = 196704

//000001000000
//000001100000
//000000100000 = 1074135072

// --
//  --
//000001100000
//000000110000 = 393264

//000000100000
//000001100000
//000001000000 = 537264192

// T
//000001110000
//000000100000 = 458784

//000000100000
//000001100000
//000000100000 = 537264160

//000000100000
//000001110000
//000000000000 = 537329664

//000000100000
//000000110000
//000000100000 = 537067552

var rotationsByType = map[PieceType][]*big.Int{
	TShape:      {big.NewInt(458784), big.NewInt(537067552), big.NewInt(537329664), big.NewInt(537264160)},
	ZigZagLeft:  {big.NewInt(196704), big.NewInt(1074135072)},
	ZigZagRight: {big.NewInt(393264), big.NewInt(537264192)},
	IShape:      {big.NewInt(240), big.NewInt(4399120515136)},
	RightLShape: {big.NewInt(524512), big.NewInt(1074004160), big.NewInt(917536), big.NewInt(3221749888)},
	LeftLShape:  {big.NewInt(131296), big.NewInt(1610743840), big.NewInt(917632), big.NewInt(1074004064)},
	SquareShape: {big.NewInt(393312)},
}

type RotationType int

const (
	Left  RotationType = -1
	Right RotationType = 1
)

type Piece struct {
	rotationCount int
	PieceType     PieceType
	rotations     []*big.Int
	field         *Field
}

func MakePiece(field *Field, pieceType PieceType) Piece {
	rotations := rotationsByType[pieceType]
	rotationsCopy := copyRotations(rotations)
	return Piece{
		rotationCount: 0,
		PieceType:     pieceType,
		rotations:     rotationsCopy,
		field:         field,
	}
}

func copyRotations(rotations []*big.Int) []*big.Int {
	rotationsCopy := make([]*big.Int, len(rotations))
	copy(rotationsCopy, rotations)
	for i, rotation := range rotationsCopy {
		rotationsCopy[i] = big.NewInt(0).Set(rotation)
	}
	return rotationsCopy
}

func (piece *Piece) Rotate(rotationType RotationType) bool {
	var diff int
	if rotationType == Left {
		diff = -1
	} else {
		diff = 1
	}

	piece.changeRotationCount(diff)
	if !piece.field.Intersects(piece.GetVal()) {
		return true
	}
	piece.changeRotationCount(-diff)
	return false
}

func (piece *Piece) changeRotationCount(diff int) {
	maxRotations := len(rotationsByType[piece.PieceType])
	piece.rotationCount += diff
	if piece.rotationCount < 0 {
		piece.rotationCount = maxRotations - 1
	} else if piece.rotationCount == maxRotations {
		piece.rotationCount = 0
	}
}

func (piece *Piece) GetVal() *big.Int {
	abs := int64(math.Abs(float64(piece.rotationCount % rotationsCntByType[piece.PieceType])))
	return piece.rotations[abs]
}

func (piece *Piece) MoveLeft() bool {
	newPieceVal := big.NewInt(0).Set(piece.GetVal())
	newPieceVal.Rsh(newPieceVal, 1)
	if piece.field.Intersects(newPieceVal) {
		return false
	}
	for i := range piece.rotations {
		newRotation := big.NewInt(0).Set(piece.rotations[i])
		piece.rotations[i] = newRotation.Rsh(newRotation, 1)
	}
	return true
}

func (piece *Piece) MoveRight() bool {
	newPieceVal := big.NewInt(0).Set(piece.GetVal())
	newPieceVal.Lsh(newPieceVal, 1)
	if piece.field.Intersects(newPieceVal) {
		return false
	}
	for i := range piece.rotations {
		newRotation := big.NewInt(0).Set(piece.rotations[i])
		piece.rotations[i] = newRotation.Lsh(newRotation, 1)
	}
	return true
}

func (piece *Piece) MoveDown() bool {
	if !piece.CanMoveDown() {
		return false
	}
	for i := range piece.rotations {
		newRotation := big.NewInt(0).Set(piece.rotations[i])
		piece.rotations[i] = newRotation.Lsh(newRotation, FieldWidth)
	}
	return true
}

func (piece *Piece) CanMoveDown() bool {
	movedDownPiece := big.NewInt(0).Lsh(piece.GetVal(), FieldWidth)
	return !piece.field.Intersects(movedDownPiece)
}
