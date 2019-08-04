package common

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

var (
	kernel32 *syscall.LazyDLL  = syscall.NewLazyDLL(`kernel32.dll`)
	proc     *syscall.LazyProc = kernel32.NewProc(`SetConsoleTextAttribute`)
)

//color code
const (
	Black = iota
	Blue
	Green
	Cyan
	Red
	Purple
	Yellow
	Light_gray
	Gray
	Light_blue
	Light_green
	Light_cyan
	Light_red
	Light_purple
	Light_yellow
	White
)

//return status code
const (
	NormalReturn = iota
	NotFound
	Eventer
	Coster
)

//################################ tool function

//printf a string with special color
func ColorPrint(i int, format string, arg ...interface{}) {
	proc.Call(uintptr(syscall.Stdout), uintptr(i))
	fmt.Printf(format, arg...)
	proc.Call(uintptr(syscall.Stdout), uintptr(15))
}

//clear up console
func ClearConsole() error {
	env := runtime.GOOS
	var cmd *exec.Cmd
	switch env {
	case "linux":
		cmd = exec.Command("clear") //Linux example, its tested
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls") //Windows example, its tested
	default:
		return fmt.Errorf("Unsuppose clear function!")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
	return nil
}

func GetTurnCode() int {
	ColorPrint(8, "choise: [ eventer:1 ]  [ coster:2 ]")
	ColorPrint(8, "\n Please input where do you want to go > ")
	choise := 0
	fmt.Scanf("%d\n", &choise)
	switch choise {
	case 1:
		return Eventer
	case 2:
		return Coster
	default:
		ColorPrint(12, "No such code!")
		return NotFound
	}
}

func PrintfColorExample() {
	for i := 0; i <= 15; i++ {
		ColorPrint(i, "=%d=", i)
	}
}
