# Go ANSI Stages

ANSI Stages is a simple package meant to emulate the stages in a cli tool where
overall output isn't important unless there is a failure. It consists of a
simple buffer abstraction that utilizes ANSI escape codes in the background.

![Demo](./.github/.img/demo.gif)

## Install

`go get github.com/louislef299/ansi`

## Examples

Create a simple buffer:

```go
// Creates a new buffer with a buffer size of 5
// Returns the buffer object
buff := New(context.TODO(), 5)
buff.Printf("hello world")
```

Erase the existing buffer:

```go
buff := New(context.TODO(), 5)

// Prints hello world 5x
for i := 0; i < 5; i++ {
    buff.Printf("hello world")
}

buff.EraseBuffer()
```

Run print statements in stages:

```go
buff := New(context.TODO(), 5)

// Stage 1
for i := 0; i < 5; i++ {
    buff.Printf("%d: hello from stage 1", i + 1)
}
// end stage
buff.NewStage("stage 1 complete!")

// Stage 2
for i := 0; i < 5; i++ {
    buff.Printf("%d: hello from stage 2", i + 1)
}
// end stage
buff.NewStage("stage 2 complete!")
```

The ANSI Buffer also implements io.Writer:

```go
log.SetOutput(New(context.TODO(), 5))
log.Println("written from test")
```

It is unreliable under stress however, as the channels that are created for
synchronization are not reliably transferred to packages like log and fmt.

## Contributing

The tests are currently visual tests and require a human to watch the output and
verify functionality. Not ideal, but that's how it is for now. Feel free to fix
it!
