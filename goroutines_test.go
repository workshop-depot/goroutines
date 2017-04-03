package goroutines

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestGo(t *testing.T) {
	result := ``

	New().
		Go(func() {
			result = `OK`
		})

	if result != `` {
		t.Error(result)
		t.Fail()
	}
}

func TestWaitStart(t *testing.T) {
	result := ``

	New().
		WaitStart().
		Go(func() {
			result = `OK`
		})

	if result != `OK` {
		t.Error(result)
		t.Fail()
	}
}

func TestWaitGo1(t *testing.T) {
	result := ``

	err := New().
		WaitGo(time.Second).
		Go(func() {
			result = `OK`
		})

	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if result != `OK` {
		t.Error(result)
		t.Fail()
	}
}

func TestWaitGo2(t *testing.T) {
	err := New().
		WaitGo(time.Millisecond * 100).
		Go(func() {
			time.Sleep(time.Second)
		})

	if err != ErrTimeout {
		t.Error(err)
		t.Fail()
	}
}

func TestWaitGo3(t *testing.T) {
	result := make(chan int, 10)

	err := New().
		WaitGo(-1).
		Go(func() {
			result <- 1
			time.Sleep(time.Millisecond * 50)
		})

	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if <-result != 1 {
		t.Fail()
	}
}

func TestRecover(t *testing.T) {
	errStr := make(chan string, 10)

	err := New().
		WaitStart().
		Recover(func(e interface{}) {
			errStr <- fmt.Sprintf("%v", e)
		}).
		Go(func() {
			panic(`1`)
		})

	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if <-errStr != `1` {
		t.Error(errStr)
		t.Fail()
	}
}

func TestAfter1(t *testing.T) {
	result := `1`

	err := New().
		WaitStart().
		After(func() {
			result += `3`
		}).
		Go(func() {
			result += `2`
		})

	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if result != `123` {
		t.Error(result)
		t.Fail()
	}
}

func TestAfter2(t *testing.T) {
	result := `1`

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
		t.Error(err)
		t.Fail()
	}
	if result != `123` {
		t.Error(result)
		t.Fail()
	}
}

func TestAfter3(t *testing.T) {
	result := `1`

	err := New().
		WaitStart().
		Recover(func(e interface{}) {}).
		After(func() {
			result += `3`
		}).
		Go(func() {
			result += `2`
			panic(`ERR`)
		})

	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if result != `12` {
		t.Error(result)
		t.Fail()
	}
}

func TestBefore1(t *testing.T) {
	result := `1`

	err := New().
		WaitStart().
		Before(func() {
			result = `0` + result
		}).
		Go(func() {
			result += `2`
		})

	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if result != `012` {
		t.Error(result)
		t.Fail()
	}
}

func TestBefore2(t *testing.T) {
	result := `1`

	err := New().
		Recover(func(e interface{}) {}).
		WaitStart().
		Before(func() {
			result = `0` + result
		}).
		Go(func() {
			result += `2`
			panic(`XYZ`)
		})

	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if result != `012` {
		t.Error(result)
		t.Fail()
	}
}

func TestWaitGroup(t *testing.T) {
	result := make(chan int, 10)
	wg := &sync.WaitGroup{}

	err1 := New().
		WaitGroup(wg).
		Go(func() {
			result <- 1
		})
	err2 := New().
		WaitGroup(wg).
		Go(func() {
			result <- 2
		})

	if err1 != nil {
		t.Error(err1)
		t.Fail()
	}
	if err2 != nil {
		t.Error(err2)
		t.Fail()
	}
	wg.Wait()
	close(result)
	sum := 0
	for v := range result {
		sum += v
	}
	if sum != 3 {
		t.Error(sum)
		t.Fail()
	}
}

func TestWithContext(t *testing.T) {
	result := make(chan int, 10000)
	ctx := context.Background()

	err := New().
		WaitStart().
		WaitGo(time.Millisecond*110).
		WithContext(ctx, func(c context.Context) {
			for {
				select {
				case <-c.Done():
					return
				case <-time.After(time.Millisecond * 10):
					result <- 1
				}
			}
		})

	if err != ErrTimeout {
		t.Error(err)
		t.Fail()
	}
	close(result)
	sum := 0
	for v := range result {
		sum += v
	}
	if sum != 10 {
		t.Error(sum)
		t.Fail()
	}
}

func TestWithContext2(t *testing.T) {
	result := make(chan int, 1000)
	ctx, cancel := context.WithCancel(context.Background())

	err := New().
		WaitStart().
		WithContext(ctx, func(c context.Context) {
			for {
				select {
				case <-c.Done():
					result <- 1
					close(result)
					return
				case <-time.After(time.Millisecond * 10):
					result <- 1
				}
			}
		})

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	<-time.After(time.Millisecond * 100)

	cancel()
	sum := 0
	for v := range result {
		sum += v
	}
	if sum != 10 {
		t.Error(sum)
		t.Fail()
	}
}

func TestWithContext3(t *testing.T) {
	result := make(chan int, 1000)
	ctx, cancel := context.WithCancel(context.Background())

	err := New().
		WaitStart().
		After(func() { result <- 1 }, true).
		Before(func() { result <- 1 }).
		Recover(func(interface{}) {
			result <- 1
			close(result)
		}).
		WithContext(ctx, func(c context.Context) {
			defer func() {
				panic(1)
			}()
			for {
				select {
				case <-c.Done():
					return
				case <-time.After(time.Millisecond * 10):
					result <- 1
				}
			}
		})

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	<-time.After(time.Millisecond * 100)

	cancel()
	sum := 0
	for v := range result {
		sum += v
	}
	if sum != 12 {
		t.Error(sum)
		t.Fail()
	}
}

func BenchmarkDefault(b *testing.B) {
	f := func() {}
	for n := 0; n < b.N; n++ {
		go f()
	}
}

func BenchmarkGo(b *testing.B) {
	f := func() {}
	for n := 0; n < b.N; n++ {
		New().
			Go(f)
	}
}
