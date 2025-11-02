package field

import (
	"math/rand"
	"testing"
)

type alwaysZeroSource struct{}

func (alwaysZeroSource) Int63() int64    { return 0 }
func (alwaysZeroSource) Seed(seed int64) {}

func BenchmarkFillBottomAndClean(b *testing.B) {
	rnd := rand.New(alwaysZeroSource{})
	f := makeBigIntField(rnd)
	f.SelectNextPiece()
	f.SelectNextPiece()

	for i := 0; i < b.N; i++ {
		for p := 0; p < 10; p++ {
			f.RotatePiece(PieceRotateRight)

			for f.MovePiece(PieceMoveLeft) {
			}
			for j := 0; j < p; j++ {
				f.MovePiece(PieceMoveRight)
			}
			for f.MovePiece(PieceMoveDown) {
			}

			f.ConcatPiece()
			f.SelectNextPiece()
		}

		f.CleanLines()
	}
}
