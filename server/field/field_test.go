package field

import (
	"flag"
	"os"
	"testing"
)

type alwaysIShapePieceSelector struct{}

func (ps alwaysIShapePieceSelector) SelectNextPiece() PieceType {
	return IShape
}

var N = flag.Int("N", 10, "Tests iterations count")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestFillBottomAndClean(t *testing.T) {
	field := MakeDefaultField(alwaysIShapePieceSelector{})

	startFieldRepr := field.String()
	expectedCleanCount := 4

	total := *N
	step := total / 10
	if step == 0 {
		step = 1
	}

	for i := 0; i < total; i++ {
		if (i+1)%(step) == 0 {
			percent := (i + 1) * 100 / total
			t.Logf("FillBottomAndClean progress %d", percent)
		}
		for p := 0; p < 10; p++ {
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
			t.Errorf("Wrong clean count %d. Expected clean count %d", field.GetCleanCount(), expectedCleanCount)
		}
		expectedCleanCount += 4
	}

	endFieldRepr := field.String()
	if startFieldRepr != endFieldRepr {
		t.Errorf("Field changed. StartRepr `%s`. End Repr `%s`", startFieldRepr, endFieldRepr)
	}
}
