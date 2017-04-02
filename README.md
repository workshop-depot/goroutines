# goroutines
This package provides utilities to perform common tasks on goroutines - waiting for a goroutine to start, timeouts, do somethting before or after a goroutine function inside the same goroutine and recover from panic. It has a simple fluent API (see tests).

For example, here we wait for the goroutine to start, also perform some action afterward in a deffered manner, so even if the goroutine panics, it would get done. Also we recover from panic and handle the error:

```go
err := New().
	WaitStart().
	Recover(func(e interface{}) {
		// error (panic) handling ...
	}).
	After(func() {
		result += `3`
	}, true).
	Go(func() {
		result += `2`
		panic(`ERR`)
	})

if err != nil {
    // ...
}
```
