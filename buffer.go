package ansi

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
)

// A Buffer represents the streaming buffer used for the ANSI stages
type Buffer struct {
	// The length of the visible output to the user
	BufferSize int

	// A prefix to print before each line
	Prefix string

	buffer []string
}

// New starts a goroutine to print or erase lines and cancels on context.Done().
// The return values
func (t *Buffer) New(ctx context.Context) (chan<- string, chan<- string) {
	printer := make(chan string)
	eraser := make(chan string)
	var w sync.Mutex

	go func() {
		for {
			select {
			case p := <-printer:
				w.Lock()
				t.Print(p)
				w.Unlock()
			case s := <-eraser:
				w.Lock()
				t.EraseBuffer()
				color.Green(s)
				t.NewStage()
				w.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()
	return printer, eraser
}

// Resets the Buffer buffer
func (t *Buffer) NewStage() {
	t.buffer = []string{}
}

// Erases the buffer size of lines
func (t *Buffer) EraseBuffer() {
	EraseLines(t.BufferSize)
}

// Print runs the logic required to actually print the output to the desired
// line in a scrolling fashion
func (t *Buffer) Print(a ...string) {
	s := strings.Join(a, " ")
	t.buffer = append(t.buffer, s)
	prefixSet := strings.Compare(t.Prefix, "") != 0

	if len(t.buffer) <= t.BufferSize {
		if prefixSet {
			fmt.Printf("%s ", t.Prefix)
		}
		fmt.Println(s)
	} else {
		EraseLines(t.BufferSize)
		for i := t.BufferSize; i > 0; i-- {
			if prefixSet {
				fmt.Printf("%s ", t.Prefix)
			}
			fmt.Println(t.buffer[len(t.buffer)-i])
		}
	}
}
