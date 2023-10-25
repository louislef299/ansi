package scroll

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
)

// A Buffer represents the streaming buffer used for the ANSI scrolling stages
type Buffer struct {
	// The length of the visible output to the user
	bufferSize int

	// A prefix to print before each line
	prefix string

	// Set the color of output text
	printerColor color.Attribute
	stageColor   color.Attribute

	w io.Writer

	// Internal synchronization variables
	buffer  []string
	eraser  chan string
	printer chan string
	stagger chan struct{}
	lock    *sync.RWMutex

	// Standard buffer synchronization requirements
	stdBuffer bool
	done      chan struct{}
}

var std = defaultBuffer()

// Default returns the standard buffer used by the package-level output functions.
func Default() *Buffer { return std }

// defaultBuffer is used to set the standard buffer internally.
func defaultBuffer() *Buffer {
	b := New(context.TODO(), os.Stdout, 15)
	b.stdBuffer = true
	return b
}

// New creates a new Buffer which starts a goroutine to print or erase lines and
// cancels on context.Done(). Returns a new Buffer to allow for scroll output to
// be written.
func New(ctx context.Context, w io.Writer, bufferSize int) *Buffer {
	b := &Buffer{
		bufferSize: bufferSize,
		w:          w,

		eraser:  make(chan string),
		lock:    &sync.RWMutex{},
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
				buff.print(buff.prefix, p)
				if b.stdBuffer {
					b.done <- struct{}{}
				}
			case e := <-b.eraser:
				buff.eraseBuffer()
				buff.buffer = []string{}
				if strings.Compare("", e) != 0 {
					buff.getColorWriter(EraserStage).Println(e)
				}
				b.done <- struct{}{}
			case <-ctx.Done():
				return
			}
		}
	}(b)
	return b
}

// print runs the logic required to actually print the output to the desired
// line in a scrolling fashion.
func (b *Buffer) print(a ...string) {
	s := strings.TrimSpace(strings.Join(a, " "))
	b.buffer = append(b.buffer, s)
	c := b.getColorWriter(PrinterStage)

	if len(b.buffer) <= b.bufferSize {
		c.Fprintln(b.w, s)
	} else {
		b.eraseBuffer()
		for i := b.bufferSize; i > 0; i-- {
			c.Fprintln(b.w, b.buffer[len(b.buffer)-i])
		}
	}
}

// Erases the buffer size of lines
func (b *Buffer) eraseBuffer() {
	if len(b.buffer) == 0 {
		return // nothing to erase
	} else if len(b.buffer) < b.bufferSize {
		b.eraseLines(len(b.buffer))
	} else {
		b.eraseLines(b.bufferSize)
	}
}

// Resets the Buffer buffer by erasing buffer output and printing out the string
// input to the screen.
func (b *Buffer) NewStage(format string, a ...interface{}) {
	if b.bufferSize == 0 {
		panic("your buffer hasn't been initialized!")
	}
	b.eraser <- fmt.Sprintf(format, a...)
	<-b.done
}

// EraseBuffer is the exported function that includes Buffer validations.
func (b *Buffer) EraseBuffer() {
	b.eraser <- ""
	<-b.done
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

func (b *Buffer) SetOutput(w io.Writer) {
	b.w = w
}

func (b *Buffer) SetPrefix(prefix string) {
	b.prefix = prefix
}

func (b *Buffer) SetPrinterColor(color color.Attribute) {
	b.printerColor = color
}

func (b *Buffer) SetStageColor(color color.Attribute) {
	b.stageColor = color
}

// Write implements io.Writer for Buffer to be used as output in other types.
// This functionality is EXPERIMENTAL. The inherent channels aren't copied over
// properly to most packages, so behavior isn't as expected.
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.stagger <- struct{}{}
	defer func() {
		<-b.stagger
	}()

	b.printer <- fmt.Sprint(strings.TrimSpace(string(p)))
	return len(p), nil
}

// EraseBuffer is the exported function that includes Buffer validations.
func EraseBuffer() {
	std.eraser <- ""
	<-std.done
}

// Resets the Buffer buffer by erasing buffer output and printing out the string
// input to the screen for the standard buffer.
func NewStage(format string, a ...interface{}) {
	std.eraser <- fmt.Sprintf(format, a...)
	<-std.done
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

func SetOutput(w io.Writer) {
	std.w = w
}

func SetPrefix(prefix string) {
	std.prefix = prefix
}

func SetPrinterColor(color color.Attribute) {
	std.printerColor = color
}

func SetStageColor(color color.Attribute) {
	std.stageColor = color
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
		c = b.stageColor
	default:
		c = b.printerColor
	}
	return color.New(c)
}
