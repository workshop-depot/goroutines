# goroutines
This package provides utilities to perform common tasks on goroutines - waiting for a goroutine to start, timeouts, do somethting before or after a goroutine function inside the same goroutine and recover from panic. It has a simple fluent API (see tests).

For example, here we wait for the goroutine to start, also perform some action afterward in a deffered manner, so even if the goroutine panics, it would get done. Also we recover from panic and handle the error:

```go
err := New().
	EnsureStarted().
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

Other patterns also can be implemented using these basic abstractions, such as a simple supervisor:

```go
func simpleSupervisor(intensity int, fn func()) {
	if intensity <= 0 {
		return
	}
	intensity--
	New().
		Recover(func(e interface{}) {
			time.Sleep(time.Millisecond * 10)
			go simpleSupervisor(intensity, fn)
		}).
		Go(fn)
}
```

Here `simpleSupervisor(...)` will restart a goroutine for a function, in case of a `panic(...)`, number of `intensity` times. And we can use it as:

```go
go simpleSupervisor(9, func() {
	// ...
})
```

This restarts the function at most, nine times and then stops. If we needed the function to run forever, we could simply drop intensity, or handle a value of `-1`.
