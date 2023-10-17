package ansi_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/fatih/color"
	. "github.com/louislef299/ansi"
)

func TestBufferStages(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	buff := &Buffer{
		BufferSize:   5,
		Prefix:       "=>",
		EraserColor:  color.FgGreen,
		PrinterColor: color.FgHiMagenta,
	}
	printer, erase := buff.New(ctx)

	for i := 0; i < 5; i++ {
		runSampleStage(printer)
		erase <- fmt.Sprintf("=>=> stage %d finished!", i)
	}
}

// Runs a sample stage to generate output
func runSampleStage(printer chan<- string) {
	stage1 := []string{
		"hello flacko",
		"hello yams",
		"hello ferg",
		"hello twelvyy",
	}

	ticker := time.Tick(time.Millisecond * 20)
	for i := 0; i < 50; i++ {
		part := i % len(stage1)
		<-ticker
		printer <- fmt.Sprintf("%d: %s", i, stage1[part])
	}
}
