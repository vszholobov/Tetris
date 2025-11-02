package keyboard

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

const showCursorASCII = "\033[?25h"
const hideCursorASCII = "\033[?25l"

type OutputController struct {
	clearFunction func()
}

func (oc *OutputController) Clear() {
	oc.clearFunction()
}

func (oc *OutputController) HideCursor() {
	fmt.Print(hideCursorASCII)
}

func (oc *OutputController) ShowCursor() {
	fmt.Print(showCursorASCII)
}

func InitOutputController() *OutputController {
	return &OutputController{clearFunction: selectClearFunction()}
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
