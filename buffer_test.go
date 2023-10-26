package scroll_test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/louislef299/scroll"
	"golang.org/x/term"
)

func TestBufferCreation(t *testing.T) {
	fmt.Println("Testing Buffer Creation:")
	b := scroll.New(context.TODO(), os.Stdout, 5)
	b.Println("hello world")
}

func TestStandardBufferTicker(t *testing.T) {
	fmt.Println("Testing the Standard Buffer:")
	scroll.SetPrefix("()=>")
	scroll.SetPrinterColor(color.FgHiRed)

	ticker := time.Tick(time.Millisecond * 20)
	for i := 0; i < 20; i++ {
		<-ticker
		scroll.Printf("%d: hello world", i)
	}
	time.Sleep(time.Second)
	scroll.EraseBuffer()
}

func TestEraseBuffer(t *testing.T) {
	fmt.Println("Testing Erase Buffer:")
	buff := scroll.New(context.TODO(), os.Stdout, 3)

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

	buff := scroll.New(ctx, os.Stdout, 4)
	buff.SetPrefix("=>")

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

	buff := scroll.New(ctx, os.Stdout, 5)
	buff.SetPrefix("=>")
	buff.SetPrinterColor(color.FgHiMagenta)
	buff.SetStageColor(color.FgGreen)

	for i := 0; i < 3; i++ {
		runSampleStage(buff, 50, time.Millisecond*20)
		buff.NewStage("=>=> stage %d finished!", i)
	}
}

func TestStandardStagesColor(t *testing.T) {
	fmt.Println("Testing Standard Buffer with Color:")

	scroll.SetPrefix("=>")
	scroll.SetPrinterColor(color.FgHiMagenta)
	scroll.SetStageColor(color.FgGreen)

	for i := 0; i < 3; i++ {
		ticker := time.Tick(time.Millisecond * 100)
		for i := 0; i < 5; i++ {
			<-ticker
			scroll.Printf("test line %d", i)
		}
		scroll.NewStage("=>=> stage %d finished!", i)
	}
}

func TestBufferStagesQuickly(t *testing.T) {
	fmt.Println("Testing Buffer Stages Quickly:")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	buff := scroll.New(ctx, os.Stdout, 5)
	buff.SetPrefix("=>")

	for i := 0; i < 5; i++ {
		runSampleStage(buff, 50, time.Millisecond*20)
		buff.NewStage("=>=> stage %d finished!", i)
	}
}

func TestStandardBufferStagesQuickly(t *testing.T) {
	fmt.Println("Testing Standard Buffer Stages Quickly:")
	scroll.SetStageColor(color.FgBlue)

	for i := 0; i < 5; i++ {
		ticker := time.Tick(time.Millisecond * 20)
		for i := 0; i < 30; i++ {
			<-ticker
			scroll.Printf("%d: testing testing", i)
		}

		scroll.NewStage("=>=> stage %d finished!", i)
	}
}

func TestEmptyErase(t *testing.T) {
	fmt.Println("Testing Empty Erase(this line should still show):")
	scroll.EraseBuffer()
}

func TestLogWriterSimple(t *testing.T) {
	log.SetOutput(scroll.New(context.TODO(), os.Stdout, 5))
	log.Println("written from test")
}

func TestLogWriterStage(t *testing.T) {
	fmt.Println("Testing Buffer Log Stages:")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	buff := scroll.New(ctx, os.Stdout, 5)
	buff.SetPrefix("=>")
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
func runSampleStage(b *scroll.Buffer, iterations int, wait time.Duration) {
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

func TestStressBuffer(t *testing.T) {
	fmt.Println("Attempting to stress the Buffer(all lines should get erased after pause)")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	buff := scroll.New(ctx, os.Stdout, 15)

	var wg sync.WaitGroup
	routines := 10000
	wg.Add(routines)
	for i := 0; i < routines; i++ {
		go func(n int) {
			defer wg.Done()
			r := rand.Float64()
			buff.Printf("hello from %d http://long-url-%g", n, r)
		}(i)
	}
	wg.Wait()
	buff.Println("sleeping 3 seconds to realize output...")
	time.Sleep(time.Second * 3)
	buff.EraseBuffer()

}

func TestMultiLineDeletion(t *testing.T) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	fmt.Printf("height: %d\nwidth: %d\n", h, w)

	repeat := "repeat"
	var printStr string
	for i := 0; i < w+10; i += len(repeat) {
		printStr = printStr + repeat
	}
	scroll.Println(printStr)
	scroll.Println("all should be erased")
	time.Sleep(time.Second)
	scroll.EraseBuffer()
}
