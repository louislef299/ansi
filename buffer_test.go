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

func TestBufferCreation(t *testing.T) {
	fmt.Println("Testing Buffer Creation:")
	b := New(context.TODO(), 5)
	b.Println("hello world")
}

func TestEraseBuffer(t *testing.T) {
	fmt.Println("Testing Erase Buffer:")
	buff := New(context.TODO(), 3)

	ticker := time.Tick(time.Millisecond * 100)
	for i := 0; i < 5; i++ {
		<-ticker
		buff.Printf("test line %d", i)
	}
	buff.Println("erase me!")
	time.Sleep(time.Second)

	buff.EraseBuffer()
}

func TestBufferStagesSlow(t *testing.T) {
	fmt.Println("Testing Buffer Stages Slowly:")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	buff := New(ctx, 4)
	buff.Prefix = "=>"

	for i := 0; i < 2; i++ {
		runSampleStage(buff, 8, time.Millisecond*200)
		buff.NewStage("=>=> stage %d finished!", i)
	}
}

func TestBufferStagesColor(t *testing.T) {
	fmt.Println("Testing Buffer Stages with Color:")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	buff := New(ctx, 5)
	buff.Prefix = "=>"
	buff.PrinterColor = color.FgHiMagenta
	buff.StageColor = color.FgGreen

	for i := 0; i < 3; i++ {
		runSampleStage(buff, 50, time.Millisecond*20)
		buff.NewStage("=>=> stage %d finished!", i)
	}
}

func TestBufferStagesQuickly(t *testing.T) {
	fmt.Println("Testing Buffer Stages Quickly:")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	buff := New(ctx, 5)
	buff.Prefix = "=>"

	for i := 0; i < 5; i++ {
		runSampleStage(buff, 50, time.Millisecond*20)
		buff.NewStage("=>=> stage %d finished!", i)
	}
}

func TestLogWriterSimple(t *testing.T) {
	log.SetOutput(New(context.TODO(), 5))
	log.Println("written from test")
}

func TestLogWriterStage(t *testing.T) {
	fmt.Println("Testing Buffer Log Stages:")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	buff := New(ctx, 5)
	buff.Prefix = "=>"
	log.SetOutput(buff)

	ticker := time.Tick(time.Millisecond * 100)
	for i := 0; i < 15; i++ {
		<-ticker
		log.Printf("%d: logging", i)
	}
	buff.NewStage("successful logger")
	time.Sleep(time.Second)
}

// Runs a sample stage to generate output
func runSampleStage(b *Buffer, iterations int, wait time.Duration) {
	stage1 := []string{
		"hello flacko",
		"hello yams",
		"hello ferg",
		"hello twelvyy",
	}

	ticker := time.Tick(wait)
	for i := 0; i < iterations; i++ {
		part := i % len(stage1)
		<-ticker
		b.Printf("%d: %s", i, stage1[part])
	}
}
