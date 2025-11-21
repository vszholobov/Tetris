package field

import (
	"crypto/rand"
	"flag"
	"math/big"
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

var N = flag.Int("N", 10000, "Tests iterations count")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestFillBottomAndClean(t *testing.T) {
	field := MakeDefaultField(fixedPieceSelector{pieceType: lib.IShape})

	startFieldRepr := field.Bytes()
	expectedCleanCount := uint16(4)

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

	endFieldRepr := field.Bytes()
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

	for pieceType := lib.PieceType(0); uint8(pieceType) < lib.PieceTypeCount; pieceType++ {
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

func bigIntToBytes32(bi *big.Int) [32]byte {
	var arr [32]byte
	b := bi.Bytes()
	copy(arr[32-len(b):], b)
	return arr
}

func randomBigInt252() *big.Int {
	randBits := new(big.Int)
	max := new(big.Int).Lsh(big.NewInt(1), 251)
	max.Sub(max, big.NewInt(1))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}
	randBits.Set(n)

	firstBit := new(big.Int).Lsh(big.NewInt(1), 251)
	randBits.Or(randBits, firstBit)
	return randBits
}

func bigIntToBinStr252(bi *big.Int) string {
	s := bi.Text(2)
	if len(s) < 256 {
		s = strings.Repeat("0", 256-len(s)) + s
	}
	return s
}

func TestMakeFieldRandom252Bits(t *testing.T) {
	total := *N
	step := total / 10
	if step == 0 {
		step = 1
	}

	for i := 0; i < total; i++ {
		if (i+1)%(step) == 0 {
			percent := (i + 1) * 100 / total
			t.Logf("TestMakeFieldRandom252Bits progress %d", percent)
		}

		n := randomBigInt252()
		binStr := bigIntToBinStr252(n)
		field := MakeField(fixedPieceSelector{pieceType: lib.IShape}, binStr)

		pieceVal := big.NewInt(240)
		b := field.Bytes()
		fieldAndPiece := big.NewInt(0).Or(big.NewInt(0).SetBytes(b[:]), pieceVal)
		expected := bigIntToBytes32(fieldAndPiece)
		actual := field.Bytes()

		if actual != expected {
			t.Fatalf("iter %d: Bytes() mismatch\nexpected: %v\n actual: %v\nbinStr: %s", i, expected, actual, binStr)
		}
	}
}
