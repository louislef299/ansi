package ansi

import (
	"fmt"
	"strings"
)

type Terminal struct {
	BufferSize int
	buffer     []string
	prefix     string
}

func New(bufferSize int) *Terminal {
	return &Terminal{
		BufferSize: bufferSize,
	}
}

// Resets the terminal buffer
func (t *Terminal) NewStage() {
	t.buffer = []string{}
}

func (t *Terminal) SetPrefix(p string) {
	t.prefix = p
}

func (t *Terminal) EraseBuffer() {
	EraseLines(t.BufferSize)
}

func (t *Terminal) Print(a ...string) {
	s := strings.Join(a, " ")
	t.buffer = append(t.buffer, s)
	prefixSet := strings.Compare(t.prefix, "") != 0

	if len(t.buffer) <= t.BufferSize {
		if prefixSet {
			fmt.Printf("%s ", t.prefix)
		}
		fmt.Println(s)
	} else {
		EraseLines(t.BufferSize)
		for i := t.BufferSize; i > 0; i-- {
			if prefixSet {
				fmt.Printf("%s ", t.prefix)
			}
			fmt.Println(t.buffer[len(t.buffer)-i])
		}
	}
}
