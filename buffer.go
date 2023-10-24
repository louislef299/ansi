package ansi

import (
	"context"
	"fmt"
	"io"
	"os"
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

	w io.Writer

	// Internal synchronization variables
	buffer  []string
	eraser  chan string
	printer chan string
	stagger chan struct{}
	lock    *sync.Mutex

	// Standard buffer synchronization requirements
	stdBuffer bool
	done      chan struct{}
}

var std = defaultBuffer()

// Default returns the standard buffer used by the package-level output functions.
func Default() *Buffer { return std }

// Used to set the standard buffer
func defaultBuffer() *Buffer {
	b := New(os.Stdout, context.TODO(), 15)
	b.stdBuffer = true
	return b
}

// New starts a goroutine to print or erase lines and cancels on contexb.Done().
// The return values
func New(w io.Writer, ctx context.Context, bufferSize int) *Buffer {
	b := &Buffer{
		BufferSize: bufferSize,
		w:          w,

		eraser:  make(chan string),
		lock:    &sync.Mutex{},
		printer: make(chan string),
		stagger: make(chan struct{}, bufferSize),
		done:    make(chan struct{}),
	}

	go func(buff *Buffer) {
		defer close(b.printer)
		defer close(buff.eraser)
		defer close(b.stagger)
		defer close(b.done)

		for {
			select {
			case p := <-b.printer:
				buff.print(p)
				if b.stdBuffer {
					b.done <- struct{}{}
				}
			case e := <-b.eraser:
				buff.eraseBuffer()
				buff.buffer = []string{}
				if strings.Compare("", e) != 0 {
					buff.getColorWriter(EraserStage).Println(e)
				}
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
	b.stagger <- struct{}{}
	defer func() {
		<-b.stagger
	}()

	b.printer <- fmt.Sprintf(format, a...)
}

// Println safely executes the channel printing logic and formats the provided
// string to the temporary buffer.
func (b *Buffer) Println(a ...interface{}) {
	b.stagger <- struct{}{}
	defer func() {
		<-b.stagger
	}()

	b.printer <- fmt.Sprint(a...)
}

// print runs the logic required to actually print the output to the desired
// line in a scrolling fashion.
func (b *Buffer) print(a ...string) {
	b.lock.Lock()
	defer b.lock.Unlock()
	s := strings.Join(a, " ")
	b.buffer = append(b.buffer, s)
	prefixSet := strings.Compare(b.Prefix, "") != 0
	c := b.getColorWriter(PrinterStage)

	if len(b.buffer) <= b.BufferSize {
		if prefixSet {
			c.Fprintf(b.w, "%s ", b.Prefix)
		}
		c.Fprintln(b.w, s)
	} else {
		b.eraseBuffer()
		for i := b.BufferSize; i > 0; i-- {
			if prefixSet {
				c.Fprintf(b.w, "%s ", b.Prefix)
			}
			c.Fprintln(b.w, b.buffer[len(b.buffer)-i])
		}
	}
}

// Write implements io.Writer for Buffer to be used as output in other types.
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.stagger <- struct{}{}
	defer func() {
		<-b.stagger
	}()

	b.printer <- fmt.Sprint(strings.TrimSpace(string(p)))
	return len(p), nil
}

// Printf safely executes the channel printing logic and formats the provided
// string to the standard buffer.
func Printf(format string, a ...interface{}) {
	std.stagger <- struct{}{}
	defer func() {
		<-std.stagger
	}()

	std.printer <- fmt.Sprintf(format, a...)
	<-std.done
}

// Println safely executes the channel printing logic and formats the provided
// string to the standard buffer.
func Println(a ...interface{}) {
	std.stagger <- struct{}{}
	defer func() {
		<-std.stagger
	}()

	std.printer <- fmt.Sprint(a...)
	<-std.done
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
