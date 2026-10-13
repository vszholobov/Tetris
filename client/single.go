package main

import (
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"
	"tetrisClient/keyboard"
	"time"

	"github.com/mattn/go-tty"
	lib "github.com/vszholobov/tetrisLib"
	"github.com/vszholobov/tetrisLib/single"
)

func runSingle(args []string) error {
	flags := flag.NewFlagSet("--single", flag.ContinueOnError)
	seed := flags.Int64("seed", time.Now().UnixNano(), "piece sequence seed")
	trace := flags.Bool("trace", false, "log input and gravity for integration tests")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	terminal, err := tty.Open()
	if err != nil {
		return fmt.Errorf("open terminal: %w", err)
	}
	defer terminal.Close()
	// Disable ISIG too: Ctrl+C is an input byte, including on kernels without signals.
	if _, err := terminal.Raw(); err != nil {
		return fmt.Errorf("raw terminal: %w", err)
	}
	if err := keyboard.ClearScreen(); err != nil {
		return fmt.Errorf("clear screen: %w", err)
	}
	fmt.Print("\x1b[?25l")
	defer fmt.Print("\x1b[?25h")

	// A bounded queue lets the reader accept a short burst while a frame is written.
	keys := make(chan rune, 64)
	errors := make(chan error, 1)
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			key, err := terminal.ReadRune()
			if err != nil {
				errors <- err
				return
			}
			select {
			case keys <- key:
			case <-done:
				return
			}
		}
	}()
	return playSingle(os.Stdout, single.New(*seed), keys, errors, *trace)
}

func playSingle(out io.Writer, game *single.Game, keys <-chan rune, errors <-chan error, trace bool) error {
	var renderer singleRenderer
	if err := renderer.Render(out, game.State()); err != nil {
		return err
	}
	if trace {
		fmt.Fprint(out, "\r\nTETRIS READY\r\n")
	}
	timer := time.NewTimer(game.Interval())
	defer timer.Stop()
	frames := time.NewTicker(time.Second / 30)
	defer frames.Stop()
	dirty := false
	for !game.Over() {
		select {
		case err := <-errors:
			return fmt.Errorf("read terminal: %w", err)
		case key := <-keys:
			if key == 'x' || key == 27 || key == 3 {
				fmt.Fprintf(out, "\x1b[%d;1H\r\nTETRIS EXIT\r\n", lib.FieldHeight+2)
				return nil
			}
			changed := game.Input(key)
			dirty = dirty || changed
			if trace {
				fmt.Fprintf(out, "\r\nTETRIS INPUT key=%c changed=%t\r\n", key, changed)
			}
		case <-timer.C:
			game.Tick()
			// Rendering no longer adds its duration to the gravity interval.
			timer.Reset(game.Interval())
			dirty = true
			if trace {
				fmt.Fprint(out, "\r\nTETRIS TICK\r\n")
			}
		case <-frames.C:
			if dirty {
				if err := renderer.Render(out, game.State()); err != nil {
					return err
				}
				dirty = false
			}
		}
	}
	if err := renderer.Render(out, game.State()); err != nil {
		return err
	}
	fmt.Fprintf(out, "\x1b[%d;1H\r\nGAME OVER score=%d\r\n", lib.FieldHeight+2, game.State().Score)
	return nil
}

type singleRenderer struct {
	previous [lib.FieldHeight + 2]string
}

func singleRows(state lib.GameStateMessage) [lib.FieldHeight + 2]string {
	var rows [lib.FieldHeight + 2]string
	rows[0] = "Tetris | a/d: move, s: down, q/e: rotate, x: quit"
	bits := new(big.Int).SetBytes(state.FieldBytes[:])
	for row := 0; row < lib.FieldHeight; row++ {
		var line strings.Builder
		for col := lib.FieldWidth - 1; col >= 0; col-- {
			if bits.Bit(row*lib.FieldWidth+col) != 0 {
				line.WriteByte('#')
			} else {
				line.WriteByte(' ')
			}
		}
		if row == 0 {
			line.WriteString("    Next:")
		} else if row == 1 || row == 2 {
			preview := RepresentationByType[state.NextPiece]
			line.WriteString("    ")
			if row-1 < len(preview) {
				line.WriteString(strings.ReplaceAll(preview[row-1], "Ж", "#"))
			}
		}
		rows[row+1] = line.String()
	}
	rows[len(rows)-1] = fmt.Sprintf("Score: %d | Speed: %d | Lines: %d", state.Score, state.Speed, state.CleanCount)
	return rows
}

// Render only changed spans. Erase stale cells with spaces, never ESC[K to the
// far edge of the terminal. A failed write must not advance the frame cache.
func (r *singleRenderer) Render(out io.Writer, state lib.GameStateMessage) error {
	rows := singleRows(state)
	var frame strings.Builder
	for row, next := range rows {
		previous := r.previous[row]
		width := max(len(next), len(previous))
		next += strings.Repeat(" ", width-len(next))
		previous += strings.Repeat(" ", width-len(previous))
		start := 0
		for start < width && next[start] == previous[start] {
			start++
		}
		end := width
		for end > start && next[end-1] == previous[end-1] {
			end--
		}
		if start < end {
			fmt.Fprintf(&frame, "\x1b[%d;%dH%s", row+1, start+1, next[start:end])
		}
	}
	if frame.Len() != 0 {
		n, err := io.WriteString(out, frame.String())
		if err != nil {
			return fmt.Errorf("draw frame: %w", err)
		}
		if n != frame.Len() {
			return io.ErrShortWrite
		}
	}
	r.previous = rows
	return nil
}
