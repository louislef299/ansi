# Functions to Save

```go
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
```
