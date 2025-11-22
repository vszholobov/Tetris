package keyboard

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"

	"github.com/mattn/go-tty"
)

const showCursorASCII = "\033[?25h"
const hideCursorASCII = "\033[?25l"

type Keyboard struct {
	keyboardTransferChannel chan rune
	keyboardInputChannel    *tty.TTY
	clearFunction           func()
	closed                  atomic.Bool
}

func MakeKeyboard() *Keyboard {
	keyboardTransferChannel := make(chan rune)
	keyboardInputChannel, _ := tty.Open()
	clearFunction := selectClearFunction()
	return &Keyboard{
		keyboardTransferChannel: keyboardTransferChannel,
		keyboardInputChannel:    keyboardInputChannel,
		clearFunction:           clearFunction,
	}
}

func (keyboard *Keyboard) ProcessKeyboardInput() {
	for {
		r, _ := keyboard.keyboardInputChannel.ReadRune()
		keyboard.keyboardTransferChannel <- r
	}
}

func (keyboard *Keyboard) Read() rune {
	return <-keyboard.keyboardTransferChannel
}

func (keyboard *Keyboard) Close() {
	if keyboard.closed.CompareAndSwap(false, true) {
		keyboard.ShowCursor()
		keyboard.Clear()
		keyboard.keyboardInputChannel.Close()
	}
}

func (oc *Keyboard) Clear() {
	oc.clearFunction()
}

func (oc *Keyboard) HideCursor() {
	fmt.Print(hideCursorASCII)
}

func (oc *Keyboard) ShowCursor() {
	fmt.Print(showCursorASCII)
}

func selectClearFunction() func() {
	switch runtime.GOOS {
	case "linux":
		return func() {
			cmd := exec.Command("clear") //Linux example, its tested
			cmd.Stdout = os.Stdout
			cmd.Run()
		}
	case "darwin":
		return func() {
			cmd := exec.Command("clear") //Macos example, its tested
			cmd.Stdout = os.Stdout
			cmd.Run()
		}
	case "windows":
		return func() {
			cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
			cmd.Stdout = os.Stdout
			cmd.Run()
		}
	default:
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}
