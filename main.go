package main

import (
	"fmt"
	"log"
	"os"
	"time"

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
	//runStages(height)

	fmt.Printf("the terminal width is %d and the height is %d\n", width, height)
}

func runTicker() {
	fmt.Println("program start")
	ticker := time.Tick(time.Second)
	for i := 1; i <= 5; i++ {
		<-ticker
		fmt.Printf("\rOn %d/5", i)
	}
	fmt.Printf("\rthis should be deleted")
	fmt.Printf("\rAll is said and done.\n")
	fmt.Printf("\033[1A")
}

func runStages(bufSize int) {
	stage1 := []string{
		"hello louis",
		"hello joe",
		"hello cash",
		"hello zach",
	}

	buff := buffer{
		bufferSize: bufSize - 2,
	}

	ticker := time.Tick(time.Millisecond * 200)
	for i := 0; i < 50; i++ {
		part := i % len(stage1)
		<-ticker
		buff.print(stage1[part])
	}

	time.Sleep(time.Second)
	eraseLines(buff.bufferSize)
	fmt.Println("stage one finished!")
}

func (b *buffer) print(output string) {
	b.buffer = append(b.buffer, output)
	if len(b.buffer) <= b.bufferSize {
		fmt.Printf("%s\n", output)
	} else {
		eraseLines(b.bufferSize)
		for i := b.bufferSize; i > 0; i-- {
			fmt.Printf("%s\r", b.buffer[len(b.buffer)-i])
			fmt.Println()
		}
	}
}

func eraseLines(lines int) {
	for i := 1; i <= lines; i++ {
		fmt.Printf("\033[%dA\r", 1)
		fmt.Printf("\033[2K")
	}
}
