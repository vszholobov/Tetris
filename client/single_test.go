package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	lib "github.com/vszholobov/tetrisLib"
	"github.com/vszholobov/tetrisLib/single"
)

func TestRenderSingle(t *testing.T) {
	lines := singleRows(single.New(1).State())
	if len(lines) != 23 {
		t.Fatalf("unexpected frame height: %d", len(lines))
	}
	if lines[21] != "############" {
		t.Fatalf("floor missing: %q", lines[21])
	}
	for _, line := range lines[1:21] {
		if len(line) < 12 || line[0] != '#' || line[11] != '#' {
			t.Fatalf("invalid board row: %q", line)
		}
	}
	if !strings.Contains(lines[22], "Score: 0") {
		t.Fatal("score missing")
	}
}

func TestSinglePreviewAllPieces(t *testing.T) {
	for piece := range RepresentationByType {
		state := single.New(1).State()
		state.NextPiece = piece
		rows := singleRows(state)
		if !strings.Contains(rows[1], "Next:") {
			t.Fatal("preview label missing")
		}
		preview := rows[2][lib.FieldWidth:] + rows[3][lib.FieldWidth:]
		if strings.Count(preview, "#") != 4 {
			t.Fatalf("piece %d: expected four blocks: %q", piece, preview)
		}
	}
}

func TestSingleDirtySpansAndUnchangedFrame(t *testing.T) {
	game := single.New(2)
	var renderer singleRenderer
	var output bytes.Buffer
	if err := renderer.Render(&output, game.State()); err != nil {
		t.Fatal(err)
	}
	initial := output.Len()
	output.Reset()
	if err := renderer.Render(&output, game.State()); err != nil {
		t.Fatal(err)
	}
	if output.Len() != 0 {
		t.Fatal("unchanged frame wrote output")
	}
	if !game.Input('a') {
		t.Fatal("test move did not change the field")
	}
	if err := renderer.Render(&output, game.State()); err != nil {
		t.Fatal(err)
	}
	if output.Len() >= 100 || output.Len() >= initial/3 {
		t.Fatalf("move redraw too large: %d bytes (initial %d)", output.Len(), initial)
	}
	if strings.Contains(output.String(), "\x1b[K") {
		t.Fatal("cleared full terminal width")
	}
	t.Logf("initial frame %d bytes, horizontal move %d bytes", initial, output.Len())
}

type shortFrameWriter struct{}

func (shortFrameWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }

func TestSingleFailedWriteDoesNotCacheFrame(t *testing.T) {
	var renderer singleRenderer
	state := single.New(2).State()
	if err := renderer.Render(shortFrameWriter{}, state); err != io.ErrShortWrite {
		t.Fatalf("expected short write: %v", err)
	}
	var retry, fresh bytes.Buffer
	if err := renderer.Render(&retry, state); err != nil {
		t.Fatal(err)
	}
	if err := new(singleRenderer).Render(&fresh, state); err != nil {
		t.Fatal(err)
	}
	if retry.String() != fresh.String() {
		t.Fatal("failed frame was cached")
	}
}

func TestSingleDeltasMatchCompleteFrames(t *testing.T) {
	var screen [lib.FieldHeight + 2][80]byte
	for row := range screen {
		for col := range screen[row] {
			screen[row][col] = ' '
		}
	}
	var renderer singleRenderer
	game := single.New(2)
	for step := 0; step < 100; step++ {
		game.Input([]rune("adsqes")[step%6])
		if step%3 == 0 {
			game.Tick()
		}
		var output bytes.Buffer
		if err := renderer.Render(&output, game.State()); err != nil {
			t.Fatal(err)
		}
		text := output.String()
		row, col := 0, 0
		for len(text) > 0 {
			if text[0] == 27 {
				end := strings.IndexByte(text, 'H')
				if end < 0 {
					t.Fatalf("invalid cursor sequence %q", text)
				}
				if _, err := fmt.Sscanf(text[:end+1], "\x1b[%d;%dH", &row, &col); err != nil {
					t.Fatal(err)
				}
				row--
				col--
				text = text[end+1:]
			} else {
				screen[row][col] = text[0]
				col++
				text = text[1:]
			}
		}
		for row, expected := range singleRows(game.State()) {
			actual := strings.TrimRight(string(screen[row][:]), " ")
			if actual != strings.TrimRight(expected, " ") {
				t.Fatalf("step %d row %d: got %q, want %q", step, row, actual, expected)
			}
		}
	}
}

func TestSingleHeldKeyDoesNotStarveGravityOrExit(t *testing.T) {
	keys := make(chan rune, 64)
	done := make(chan struct{})
	go func() {
		defer close(done)
		until := time.NewTimer(600 * time.Millisecond)
		defer until.Stop()
		for {
			select {
			case keys <- 'a':
			case <-until.C:
				keys <- 'x'
				return
			}
		}
	}()
	game := single.New(2)
	initial := game.State()
	result := make(chan error, 1)
	go func() { result <- playSingle(io.Discard, game, keys, nil, false) }()
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("held key starved quit")
	}
	<-done
	// Move the original piece to the same wall without advancing time.
	withoutGravity := single.New(2)
	for withoutGravity.Input('a') {
	}
	if game.State().FieldBytes == withoutGravity.State().FieldBytes || game.State() == initial {
		t.Fatal("held key starved gravity")
	}
}

func TestSingleInvalidFlagsWithoutTTY(t *testing.T) {
	if err := runSingle([]string{"--seed", "not-a-number"}); err == nil {
		t.Fatal("invalid seed accepted")
	}
	if err := runSingle([]string{"extra"}); err == nil {
		t.Fatal("unexpected argument accepted")
	}
}
