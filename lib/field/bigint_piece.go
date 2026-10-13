package field

import (
	"math/big"

	lib "github.com/vszholobov/tetrisLib"
)

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

var rotationsByType = map[lib.PieceType][]*big.Int{
	lib.TShape:      {big.NewInt(458784), big.NewInt(537067552), big.NewInt(537329664), big.NewInt(537264160)},
	lib.ZigZagLeft:  {big.NewInt(196704), big.NewInt(1074135072)},
	lib.ZigZagRight: {big.NewInt(393264), big.NewInt(537264192)},
	lib.IShape:      {big.NewInt(240), big.NewInt(4399120515136)},
	lib.RightLShape: {big.NewInt(524512), big.NewInt(1074004160), big.NewInt(917536), big.NewInt(3221749888)},
	lib.LeftLShape:  {big.NewInt(131296), big.NewInt(1610743840), big.NewInt(917632), big.NewInt(1074004064)},
	lib.SquareShape: {big.NewInt(393312)},
}

type bigIntPiece struct {
	rotationCount int
	PieceType     lib.PieceType
	rotations     []*big.Int
	field         *BigIntField
}

func makePiece(field *BigIntField, pieceType lib.PieceType) bigIntPiece {
	rotations := rotationsByType[pieceType]
	rotationsCopy := copyRotations(rotations)
	return bigIntPiece{
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

func (piece *bigIntPiece) rotate(rotationType RotationType) bool {
	diff := int(rotationType)
	piece.changeRotationCount(diff)
	if piece.field.intersects(piece.getVal()) {
		// cancel rotation if intersects
		piece.changeRotationCount(-diff)
		return false
	}
	return true
}

func (piece *bigIntPiece) changeRotationCount(diff int) {
	maxRotations := len(rotationsByType[piece.PieceType])
	piece.rotationCount += diff
	if piece.rotationCount < 0 {
		piece.rotationCount = maxRotations - 1
	} else if piece.rotationCount == maxRotations {
		piece.rotationCount = 0
	}
}

func (piece *bigIntPiece) move(moveDirection PieceMoveDirection) bool {
	if !piece.canMove(moveDirection) {
		return false
	}

	for i := range piece.rotations {
		newRotation := big.NewInt(0).Set(piece.rotations[i])
		switch moveDirection {
		case PieceMoveLeft:
			piece.rotations[i] = newRotation.Rsh(newRotation, 1)
		case PieceMoveRight:
			piece.rotations[i] = newRotation.Lsh(newRotation, 1)
		case PieceMoveDown:
			piece.rotations[i] = newRotation.Lsh(newRotation, lib.FieldWidth)
		}
	}
	return true
}

func (piece *bigIntPiece) canMove(moveDirection PieceMoveDirection) bool {
	newPieceVal := big.NewInt(0).Set(piece.getVal())

	if piece.field.intersects(newPieceVal) {
		return false
	}

	switch moveDirection {
	case PieceMoveLeft:
		newPieceVal.Rsh(newPieceVal, 1)
	case PieceMoveRight:
		newPieceVal.Lsh(newPieceVal, 1)
	case PieceMoveDown:
		newPieceVal.Lsh(newPieceVal, lib.FieldWidth)
	default:
		return false
	}

	return !piece.field.intersects(newPieceVal)
}

// Get current rotated piece value
func (piece *bigIntPiece) getVal() *big.Int {
	idx := piece.rotationCount % rotationsCntByType[piece.PieceType]
	if idx < 0 {
		idx += rotationsCntByType[piece.PieceType]
	}
	return piece.rotations[idx]
}
