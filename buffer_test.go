package ansi_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"testing"
	"time"

	. "github.com/louislef299/ansi"
)

func TestBuffer(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	buff := &Buffer{
		BufferSize: 5,
		Prefix:     "=>",
	}
	printer, erase := buff.New(ctx)

	for i := 0; i < 5; i++ {
		runSampleStage(printer)
		erase <- fmt.Sprintf("=>=> stage %d finished!\n", i)
	}
}

func runSampleStage(printer chan<- string) {
	stage1 := []string{
		"hello john",
		"hello ringo",
		"hello george",
		"hello paul",
	}

	ticker := time.Tick(time.Millisecond * 20)
	for i := 0; i < 50; i++ {
		part := i % len(stage1)
		<-ticker
		printer <- stage1[part]
	}
}
