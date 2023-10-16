package ansi

import "fmt"

func CursorUp(line int) {
	fmt.Printf("\033[%dA", line)
}

func CursorDown(line int) {
	fmt.Printf("\033[%dB", line)
}

func CursorTo(line int) {
	fmt.Printf("\033[%d;1H", line)
}

func ClearEntireLine() {
	fmt.Printf("\033[2K")
}

func EraseLines(lines int) {
	for i := 1; i <= lines; i++ {
		CursorUp(1)
		ClearEntireLine()
	}
}
