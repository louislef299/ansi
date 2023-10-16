package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
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

	//runTicker()

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
	buff.SetPrefix("=>")

	for i := 0; i < 10; i++ {
		buff.NewStage()
		runSampleStage(buff)
		time.Sleep(time.Second)
		buff.EraseBuffer()
		color.Green("=>=> stage %d finished!\n", i)
		time.Sleep(time.Second)
	}
}

func runSampleStage(buff *Terminal) {
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
		buff.Print(stage1[part])
	}
}
