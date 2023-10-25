package scroll

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

// A Buffer represents the streaming buffer used for the ANSI scrolling stages
type Buffer struct {
	// The length of the visible output to the user
	bufferSize int

	// A prefix to print before each line
	prefix string

	// States whether the current fd is a Terminal
	isTerm bool

	// Set the color of output text
	printerColor color.Attribute
	stageColor   color.Attribute

	w      io.Writer
	buffer []string

	// Internal synchronization variables
	eraser  chan string
	printer chan string
	stagger chan struct{}

	lock *sync.RWMutex
	ctx  context.Context

	// Standard buffer synchronization requirements
	stdBuffer bool
	done      chan struct{}
}

var (
	std, cancel = defaultBuffer()

	// IsTerm dynamically prevents usage of ANSI escape sequences if stdout's
	// file descriptor is not a Terminal. NO_TERMINAL_CHECK is an environment
	// variable used to override the default terminal check in the case it is
	// incorrect.
	IsTerm = (isatty.IsTerminal(os.Stdout.Fd()) ||
		isatty.IsCygwinTerminal(os.Stdout.Fd()) ||
		os.Getenv("NO_TERMINAL_CHECK") != "")
)

// Default returns the standard buffer used by the package-level output functions.
func Default() *Buffer { return std }

// defaultBuffer is used to set the standard buffer internally.
func defaultBuffer() (*Buffer, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.TODO())
	b := New(ctx, os.Stdout, 15)
	b.stdBuffer = true
	return b, cancel
}

// New creates a new Buffer which starts a goroutine to print or erase lines and
// cancels on context.Done(). Returns a new Buffer to allow for scroll output to
// be written.
func New(ctx context.Context, w io.Writer, bufferSize int) *Buffer {
	b := &Buffer{
		eraser:  make(chan string),
		printer: make(chan string),
		stagger: make(chan struct{}, bufferSize),
		done:    make(chan struct{}),
		isTerm:  IsTerm,

		lock: &sync.RWMutex{},
		ctx:  ctx,
	}

	b.SetOutput(w)
	b.SetBufferSize(bufferSize)

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
			case <-b.ctx.Done():
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

// SetBufferSize sets the size of the Buffer.
func (b *Buffer) SetBufferSize(size int) {
	b.bufferSize = size
}

// SetOutput sets the destination output for the Buffer.
func (b *Buffer) SetOutput(w io.Writer) {
	b.w = w
}

// SetPrefix sets the prefix for output from the Buffer.
func (b *Buffer) SetPrefix(prefix string) {
	b.prefix = prefix
}

// SetPrinterColor sets the output color for scrolling output on the Buffer.
func (b *Buffer) SetPrinterColor(color color.Attribute) {
	b.printerColor = color
}

// SetStageColor sets the output color for stage finalizer output on the Buffer.
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

// CancelBuffer safely abandons work and closes the Buffer. After Cancelling a
// Buffer, it is no longer useable.
func CancelBuffer() {
	cancel()
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

// SetBufferSize sets the buffer size of the standard Buffer.
func SetBufferSize(size int) {
	std.bufferSize = size
}

// SetContext sets the context of the standard Buffer.
func SetContext(ctx context.Context) {
	std.ctx = ctx
}

// SetOutput sets the destination output for the standard Buffer.
func SetOutput(w io.Writer) {
	std.w = w
}

// SetPrefix sets the prefix for output from the standard Buffer.
func SetPrefix(prefix string) {
	std.prefix = prefix
}

// SetPrinterColor sets the output color for scrolling output on the standard
// Buffer.
func SetPrinterColor(color color.Attribute) {
	std.printerColor = color
}

// SetStageColor sets the output color for stage finalizer output on the
// standard Buffer.
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
