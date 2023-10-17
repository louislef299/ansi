package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	. "github.com/louislef299/ansi"
	"golang.org/x/term"
)

type buffer struct {
	bufferSize int
	buffer     []string
}

func main() {
	_, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	buff := &Buffer{
		BufferSize: height - 5,
		Prefix:     "=>",
	}
	printer, erase := buff.New(ctx)

	for i := 0; i < 5; i++ {
		runSampleStage(printer)
		erase <- fmt.Sprintf("=>=> stage %d finished!\n", i)
	}

	//fmt.Printf("the terminal width is %d and the height is %d\n", width, height)
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
