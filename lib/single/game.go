// Package single runs the shared Tetris field without a network session.
package single

import (
	"math/rand"
	"time"

	lib "github.com/vszholobov/tetrisLib"
	"github.com/vszholobov/tetrisLib/field"
)

type Game struct {
	field field.Field
	over  bool
}

func New(seed int64) *Game {
	return &Game{field: field.MakeDefaultField(field.MakePieceSelector(rand.New(rand.NewSource(seed))))}
}

func (g *Game) Over() bool { return g.over }

func (g *Game) Interval() time.Duration {
	return time.Second / 4 / time.Duration(g.field.GetSpeed())
}

func (g *Game) State() lib.GameStateMessage {
	result := lib.Ongoing
	if g.over {
		result = lib.Lose
	}
	return lib.GameStateMessage{
		FieldType: lib.Self, GameResult: result, FieldBytes: g.field.Bytes(),
		Speed: g.field.GetSpeed(), Score: g.field.GetScore(),
		CleanCount: g.field.GetCleanCount(), NextPiece: g.field.GetNextPieceType(),
	}
}

// Tick advances gravity, locks landed pieces and clears rows before spawning.
func (g *Game) Tick() {
	if g.over || g.field.MovePiece(field.PieceMoveDown) {
		return
	}
	g.field.ConcatPiece()
	g.field.CleanLines()
	g.field.SelectNextPiece()
	g.over = !g.field.CanMovePiece(field.PieceMoveDown)
}

func (g *Game) Input(key rune) bool {
	if g.over {
		return false
	}
	switch key {
	case 'a':
		// The field stores horizontal coordinates from the least significant bit.
		return g.field.MovePiece(field.PieceMoveRight)
	case 'd':
		return g.field.MovePiece(field.PieceMoveLeft)
	case 's':
		before := g.State()
		g.Tick()
		return before != g.State()
	case 'q':
		return g.field.RotatePiece(field.PieceRotateLeft)
	case 'e':
		return g.field.RotatePiece(field.PieceRotateRight)
	}
	return false
}
