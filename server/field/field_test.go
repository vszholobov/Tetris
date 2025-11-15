package field

import (
	"flag"
	"os"
	"strings"
	"testing"

	lib "github.com/vszholobov/tetrisLib"
)

type fixedPieceSelector struct {
	pieceType lib.PieceType
}

func (ps fixedPieceSelector) SelectNextPiece() lib.PieceType {
	return ps.pieceType
}

var N = flag.Int("N", 10, "Tests iterations count")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestFillBottomAndClean(t *testing.T) {
	field := MakeDefaultField(fixedPieceSelector{pieceType: lib.IShape})

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

func TestTopBorderClosed(t *testing.T) {
	var closedFieldString = strings.ReplaceAll(`
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
111111111111
`, "\n", "")

	for pieceType := lib.PieceType(0); int(pieceType) < lib.PieceTypeCount; pieceType++ {
		field := MakeField(fixedPieceSelector{pieceType: pieceType}, closedFieldString)
		for pieceMoveDirection := PieceMoveDirection(0); int(pieceMoveDirection) < PieceMoveDirectionCount; pieceMoveDirection++ {
			if field.CanMovePiece(pieceMoveDirection) {
				t.Errorf("Can move piece %d in direction %d", pieceType, pieceMoveDirection)
			}
		}
		rotatedLeft := field.RotatePiece(PieceRotateLeft)
		if rotatedLeft {
			t.Errorf("Can rotate piece %d left", pieceType)
		}
		rotatedRight := field.RotatePiece(PieceRotateRight)
		if rotatedRight {
			t.Errorf("Can rotate piece %d right", pieceType)
		}
	}

}
