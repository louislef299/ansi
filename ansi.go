package scroll

import "fmt"

// cursorUp uses an ANSI escape sequence to move the terminal's cursor position
// up provided lines.
func cursorUp(line int) {
	_, err := fmt.Printf("\033[%dA", line)
	if err != nil {
		panic(err)
	}
}

// clearEntireLine uses an ANSI escape sequence to delete the entire line of the
// terminal.
func clearEntireLine() {
	_, err := fmt.Printf("\033[2K")
	if err != nil {
		panic(err)
	}
}

// eraseLines scrolls up one line at a time from current position and clears
// each line.
func (b *Buffer) eraseLines(lines int) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if !b.isTerm {
		return
	}

	for i := 1; i <= lines; i++ {
		cursorUp(1)
		clearEntireLine()
	}
}
