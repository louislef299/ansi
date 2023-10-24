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

	// Set the color of output text
	PrinterColor color.Attribute
	StageColor   color.Attribute

	// Internal synchronization variables
	buffer  []string
	eraser  chan string
	printer chan string
}

// New starts a goroutine to print or erase lines and cancels on contexb.Done().
// The return values
func New(ctx context.Context, bufferSize int) *Buffer {
	var w sync.Mutex
	b := &Buffer{
		BufferSize: bufferSize,
		eraser:     make(chan string),
	}
	b.printer = make(chan string)

	go func(buff *Buffer) {
		defer close(b.printer)
		defer close(buff.eraser)
		for {
			select {
			case p := <-b.printer:
				w.Lock()
				buff.print(p)
				w.Unlock()
			case e := <-b.eraser:
				w.Lock()
				buff.eraseBuffer()
				buff.buffer = []string{}
				if strings.Compare("", e) != 0 {
					buff.getColorWriter(EraserStage).Println(e)
				}
				w.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}(b)
	return b
}

// Resets the Buffer buffer by erasing buffer output and printing out the string
// input to the screen.
func (b *Buffer) NewStage(format string, a ...interface{}) {
	if b.BufferSize == 0 {
		panic("your buffer hasn't been initialized!")
	}
	b.eraser <- fmt.Sprintf(format, a...)
}

// EraseBuffer is the exported function that includes Buffer validations.
func (b *Buffer) EraseBuffer() {
	b.NewStage("")
}

// Erases the buffer size of lines
func (b *Buffer) eraseBuffer() {
	if len(b.buffer) < b.BufferSize {
		eraseLines(len(b.buffer))
	} else {
		eraseLines(b.BufferSize)
	}
}

// Printf safely executes the channel printing logic and formats the provided
// string to the temporary buffer.
func (b *Buffer) Printf(format string, a ...interface{}) {
	b.printer <- fmt.Sprintf(format, a...)
}

// Println safely executes the channel printing logic and formats the provided
// string to the temporary buffer.
func (b *Buffer) Println(a ...interface{}) {
	b.printer <- fmt.Sprint(a...)
}

// print runs the logic required to actually print the output to the desired
// line in a scrolling fashion.
func (b *Buffer) print(a ...string) {
	s := strings.Join(a, " ")
	b.buffer = append(b.buffer, s)
	prefixSet := strings.Compare(b.Prefix, "") != 0
	c := b.getColorWriter(PrinterStage)

	if len(b.buffer) <= b.BufferSize {
		if prefixSet {
			c.Printf("%s ", b.Prefix)
		}
		c.Println(s)
	} else {
		b.eraseBuffer()
		for i := b.BufferSize; i > 0; i-- {
			if prefixSet {
				c.Printf("%s ", b.Prefix)
			}
			c.Println(b.buffer[len(b.buffer)-i])
		}
	}
}

// Write implements io.Writer for Buffer to be used as output in other types.
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.Println(strings.TrimSpace(string(p)))
	return len(p), nil
}

// Custom stage type for color function
type stage string

const (
	PrinterStage stage = "PRINTER"
	EraserStage  stage = "ERASER"
)

// getColorWriter gets the color set in the Buffer based on the stage.
func (b *Buffer) getColorWriter(s stage) *color.Color {
	var c color.Attribute
	switch s {
	case EraserStage:
		c = b.StageColor
	default:
		c = b.PrinterColor
	}
	return color.New(c)
}
