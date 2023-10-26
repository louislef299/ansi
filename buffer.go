package scroll

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// A Buffer represents the streaming buffer used for the ANSI scrolling stages
type Buffer struct {
	// The max length of the visible output to the user
	bufferMax int

	// Represents the current buffer size
	currentBufferSize int

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
	std = defaultBuffer()

	// IsTerm dynamically prevents usage of ANSI escape sequences if stdout's
	// file descriptor is not a Terminal. NO_TERMINAL_CHECK is an environment
	// variable used to override the default terminal check in the case it is
	// incorrect.
	IsTerm = (term.IsTerminal(int(os.Stdout.Fd())) ||
		os.Getenv("NO_TERMINAL_CHECK") != "")
)

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
		eraser:  make(chan string),
		printer: make(chan string),
		stagger: make(chan struct{}, bufferSize),
		done:    make(chan struct{}),
		isTerm:  IsTerm,

		lock: &sync.RWMutex{},
		ctx:  ctx,
	}

	b.SetOutput(w)
	b.SetBufferMax(bufferSize)

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
	// dynamically checks to see if the buffer will go beyond the width limit of
	// the terminal
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic(err)
	}
	output := chunk(strings.TrimSpace(strings.Join(a, " ")), w)

	if len(b.buffer) > b.bufferMax {
		// don't grow buffer more than needed
		b.buffer = append(b.buffer[1:], output...)
	} else {
		b.buffer = append(b.buffer, output...)
	}

	c := b.getColorWriter(PrinterStage)
	if b.currentBufferSize+len(output) <= b.bufferMax {
		for _, s := range output {
			c.Fprintln(b.w, s)
			b.currentBufferSize++
		}
	} else {
		b.eraseBuffer()
		for i := b.bufferMax; i > 0; i-- {
			c.Fprintln(b.w, b.buffer[len(b.buffer)-i])
			b.currentBufferSize++
		}
	}
}

// eraseBuffer erases all lines that are printed to the terminal for the
// existing Buffer.
func (b *Buffer) eraseBuffer() {
	if b.currentBufferSize == 0 {
		return // nothing to erase
	} else if b.currentBufferSize < b.bufferMax {
		b.eraseLines(b.currentBufferSize)
	} else {
		b.eraseLines(b.bufferMax)
	}
	b.currentBufferSize = 0
}

// NewStage resets the Buffer by erasing the buffer output and printing out the
// stage input to the screen.
func (b *Buffer) NewStage(format string, a ...interface{}) {
	if b.bufferMax == 0 {
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

// GetBufferSize returns the current bufferSize of the Buffer.
func (b *Buffer) GetBufferSize() int {
	return b.bufferMax
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

// SetBufferMax sets the size of the Buffer.
func (b *Buffer) SetBufferMax(size int) {
	b.bufferMax = size
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

// EraseBuffer is the exported function that includes Buffer validations.
func EraseBuffer() {
	std.eraser <- ""
	<-std.done
}

// GetBufferMax returns the current maximum buffer length of the standard
// Buffer.
func GetBufferMax() int {
	return std.bufferMax
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

// SetBufferMax sets the buffer size of the standard Buffer.
func SetBufferMax(size int) {
	std.bufferMax = size
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

func chunk(s string, chunkSize int) []string {
	if len(s) == 0 {
		return nil
	}
	if chunkSize >= len(s) {
		return []string{s}
	}
	var chunks []string = make([]string, 0, (len(s)-1)/chunkSize+1)
	currentLen := 0
	currentStart := 0
	for i := range s {
		if currentLen == chunkSize {
			chunks = append(chunks, s[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, s[currentStart:])
	return chunks
}
