package single

import (
	"math/big"
	"strings"
	"testing"

	lib "github.com/vszholobov/tetrisLib"
	"github.com/vszholobov/tetrisLib/field"
)

type squareSelector struct{}

func (squareSelector) SelectNextPiece() lib.PieceType { return lib.SquareShape }

func TestMovementAndWall(t *testing.T) {
	g := &Game{field: field.MakeDefaultField(squareSelector{})}
	start := g.State()
	if !g.Input('a') || g.State() == start {
		t.Fatal("left did not move")
	}
	if !g.Input('d') || g.State() != start {
		t.Fatal("right did not undo left")
	}
	for i := 0; i < 20; i++ {
		g.Input('a')
	}
	wall := g.State()
	if g.Input('a') || g.State() != wall {
		t.Fatal("crossed wall")
	}
	if !g.Input('s') {
		t.Fatal("soft drop did not move")
	}
	if g.Input('?') {
		t.Fatal("unknown key changed field")
	}
}

func TestGravityAndDeterminism(t *testing.T) {
	a, b := New(123), New(123)
	start := a.State()
	a.Tick()
	b.Tick()
	if a.State() == start {
		t.Fatal("gravity did not move")
	}
	for i := 0; i < 100; i++ {
		for _, key := range "aqds" {
			a.Input(key)
			b.Input(key)
		}
		a.Tick()
		b.Tick()
		if a.State() != b.State() {
			t.Fatal("seeded games diverged")
		}
	}
}

func TestRotationRoundTrip(t *testing.T) {
	g := New(5)
	start := g.State()
	if !g.Input('q') || !g.Input('e') || g.State() != start {
		t.Fatal("rotation round trip failed")
	}
}

func TestLineClearAndScore(t *testing.T) {
	board := "111111111111" + strings.Repeat("111110011111", 2) + strings.Repeat("100000000001", lib.FieldHeight-3)
	g := &Game{field: field.MakeField(squareSelector{}, board)}
	for i := 0; i < lib.FieldHeight; i++ {
		g.Tick()
		if g.State().CleanCount != 0 {
			break
		}
	}
	state := g.State()
	if state.CleanCount != 2 || state.Score == 0 || g.Over() {
		t.Fatalf("line clear failed: %+v", state)
	}
	bits := new(big.Int).SetBytes(state.FieldBytes[:])
	for col := 1; col < lib.FieldWidth-1; col++ {
		if bits.Bit((lib.FieldHeight-2)*lib.FieldWidth+col) != 0 {
			t.Fatal("cleared line retained cells")
		}
	}
}

func TestGameOverFreezesState(t *testing.T) {
	g := &Game{field: field.MakeDefaultField(squareSelector{})}
	for i := 0; i < 500 && !g.Over(); i++ {
		g.Tick()
	}
	if !g.Over() || g.State().GameResult != lib.Lose {
		t.Fatal("stacking did not end game")
	}
	final := g.State()
	g.Tick()
	if g.Input('a') || g.State() != final {
		t.Fatal("game changed after game over")
	}
}
