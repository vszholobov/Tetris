package field

import (
	"fmt"
	"testing"
)

type alwaysIShapePieceSelector struct{}

func (ps alwaysIShapePieceSelector) SelectNextPiece() PieceType {
	return IShape
}

func BenchmarkFillBottomAndClean(b *testing.B) {
	field := MakeField(alwaysIShapePieceSelector{})
	field.SelectNextPiece()
	field.SelectNextPiece()

	expectedCleanCount := 4
	for i := 0; i < b.N; i++ {
		for p := 0; p < 10; p++ {
			//
			field.RotatePiece(PieceRotateRight)

			for field.MovePiece(PieceMoveLeft) {
			}
			for j := 0; j < p; j++ {
				field.MovePiece(PieceMoveRight)
			}
			for field.MovePiece(PieceMoveDown) {
			}

			field.ConcatPiece()
			field.SelectNextPiece()
		}
		field.CleanLines()

		if expectedCleanCount != field.GetCleanCount() {
			b.Errorf("Wrong clean count %d. Expected clean count %d", field.GetCleanCount(), expectedCleanCount)
		}
		expectedCleanCount += 4
	}

	fmt.Println(field.String())
}
