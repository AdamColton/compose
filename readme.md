## Function Composition in Go

I did this more to see if I could than because it's a good idea. The compose.New
function can take two or more functions and compose them. It checks that the
output of one function matches the input of the next.

Getting the basic version working was pretty easy, it took a little work to get
it working correctly with variadic functions.

The return value does need to be cast to a func type to be used
```go
  rms := compose.New(squareSlice, avg, math.Sqrt).(func([]float64) float64)
  fmt.Println(rms(5,6,7)) // 6.055...
```