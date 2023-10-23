package ansi

import "fmt"

func cursorUp(line int) {
	fmt.Printf("\033[%dA", line)
}

func clearEntireLine() {
	fmt.Printf("\033[2K")
}

func eraseLines(lines int) {
	for i := 1; i <= lines; i++ {
		cursorUp(1)
		clearEntireLine()
	}
}
