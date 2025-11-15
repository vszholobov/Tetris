package field

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

func MakeField(pieceSelector PieceSelector) Field {
	gameField := makeBigIntField(pieceSelector)
	gameField.SelectNextPiece()
	gameField.SelectNextPiece()
	return gameField
}
