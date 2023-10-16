package ansi

import "fmt"

type Terminal struct {
	BufferSize int
	buffer     []string
}

func New(bufferSize int) *Terminal {
	return &Terminal{
		BufferSize: bufferSize,
	}
}

func (t *Terminal) Print(arg string) {
	t.buffer = append(t.buffer, arg)
	if len(t.buffer) <= t.BufferSize {
		fmt.Printf("%s\n", arg)
	} else {
		EraseLines(t.BufferSize)
		for i := t.BufferSize; i > 0; i-- {
			fmt.Printf("%s", t.buffer[len(t.buffer)-i])
			fmt.Println()
		}
	}
}
