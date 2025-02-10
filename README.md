# jxgen

PoC code generator for github.com/go-faster/jx.

```
goos: linux
goarch: amd64
pkg: github.com/ernado/jxgen/internal/example
cpu: AMD Ryzen 9 5950X 16-Core Processor            
BenchmarkEncoding
BenchmarkEncoding/jxgen
BenchmarkEncoding/jxgen-32         	29595885	        39.47 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncoding/goccy
BenchmarkEncoding/goccy-32         	26777535	        45.20 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncoding/stdlib
BenchmarkEncoding/stdlib-32        	11650272	        89.35 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncoding/easyjson
BenchmarkEncoding/easyjson-32      	53056629	        22.76 ns/op	       0 B/op	       0 allocs/op
PASS
```

On simple structure same as goccy, but easyjson still faster.