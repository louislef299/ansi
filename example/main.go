package main

import (
	"fmt"
	"log"
	"os"
	"time"

	. "github.com/louislef299/ansi"
	"golang.org/x/term"
)

type buffer struct {
	bufferSize int
	buffer     []string
}

func main() {
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	runStages(height)

	fmt.Printf("the terminal width is %d and the height is %d\n", width, height)

	runTicker()

	// for i := 0; i < 4; i++ {
	// 	fmt.Println(i)
	// }
	// CursorTo(2)
	// fmt.Println("to this line")
	// CursorTo(6)
}

func runTicker() {
	fmt.Println("program start")
	ticker := time.Tick(time.Second)
	for i := 1; i <= 5; i++ {
		<-ticker
		fmt.Printf("\rOn %d/5", i)
	}
	fmt.Printf("\rthis should be deleted")
	time.Sleep(time.Second)
	fmt.Printf("\rAll is said and done.\n")
	time.Sleep(time.Second)
	fmt.Printf("\033[1A")
}

func runStages(bufSize int) {
	buff := New(bufSize - 5)

	runSampleStage(buff)
	time.Sleep(time.Second)
	EraseLines(buff.BufferSize)
	fmt.Println("stage one finished!")

	runSampleStage(buff)
	time.Sleep(time.Second)
	EraseLines(buff.BufferSize)
	fmt.Println("stage two finished!")

	runSampleStage(buff)
	time.Sleep(time.Second)
	EraseLines(buff.BufferSize)
	fmt.Println("stage three finished!")
}

func runSampleStage(buff *Terminal) {
	stage1 := []string{
		"hello louis",
		"hello joe",
		"hello cash",
		"hello zach",
	}

	ticker := time.Tick(time.Millisecond * 200)
	for i := 0; i < 50; i++ {
		part := i % len(stage1)
		<-ticker
		buff.Print(stage1[part])
	}
}
