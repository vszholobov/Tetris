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

	keys := make(chan rune)
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
	game := single.New(*seed)
	renderSingle(os.Stdout, game.State())
	if *trace {
		fmt.Print("TETRIS READY\r\n")
	}
	timer := time.NewTimer(game.Interval())
	defer timer.Stop()
	for !game.Over() {
		select {
		case err := <-errors:
			return fmt.Errorf("read terminal: %w", err)
		case key := <-keys:
			if key == 'x' || key == 27 || key == 3 {
				fmt.Print("\r\nTETRIS EXIT\r\n")
				return nil
			}
			changed := game.Input(key)
			renderSingle(os.Stdout, game.State())
			if *trace {
				fmt.Printf("TETRIS INPUT key=%c changed=%t\r\n", key, changed)
			}
		case <-timer.C:
			game.Tick()
			renderSingle(os.Stdout, game.State())
			if *trace {
				fmt.Print("TETRIS TICK\r\n")
			}
			timer.Reset(game.Interval())
		}
	}
	fmt.Printf("\r\nGAME OVER score=%d\r\n", game.State().Score)
	return nil
}

func renderSingle(out io.Writer, state lib.GameStateMessage) {
	var frame strings.Builder
	frame.WriteString("\x1b[HTetris | a/d: move, s: down, q/e: rotate, x: quit\x1b[K\r\n")
	bits := new(big.Int).SetBytes(state.FieldBytes[:])
	for row := 0; row < lib.FieldHeight; row++ {
		for col := lib.FieldWidth - 1; col >= 0; col-- {
			if bits.Bit(row*lib.FieldWidth+col) != 0 {
				frame.WriteByte('#')
			} else {
				frame.WriteByte(' ')
			}
		}
		frame.WriteString("\x1b[K\r\n")
	}
	fmt.Fprintf(&frame, "Score: %d | Speed: %d | Lines: %d | Next: %d\x1b[K\r\n", state.Score, state.Speed, state.CleanCount, state.NextPiece)
	io.WriteString(out, frame.String())
}
