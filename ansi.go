package ansi

import "fmt"

func cursorUp(line int) {
	_, err := fmt.Printf("\033[%dA", line)
	if err != nil {
		panic(err)
	}
}

func clearEntireLine() {
	_, err := fmt.Printf("\033[2K")
	if err != nil {
		panic(err)
	}
}

func (b *Buffer) eraseLines(lines int) {
	b.lock.Lock()
	defer b.lock.Unlock()
	for i := 1; i <= lines; i++ {
		cursorUp(1)
		clearEntireLine()
	}
}
