package goroutines

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGo(t *testing.T) {
	result := ""

	New().
		Go(func() {
			result = "OK"
		})

	assert.Equal(t, "", result)
}

func TestWaitStart(t *testing.T) {
	result := ""

	New().
		EnsureStarted().
		Go(func() {
			result = "OK"
		})

	assert.Equal(t, "OK", result)
}

func TestWaitGo1(t *testing.T) {
	result := ""

	err := New().
		Timeout(time.Second).
		Go(func() {
			result = "OK"
		})

	assert.Nil(t, err)
	assert.Equal(t, "OK", result)
}

func TestWaitGo2(t *testing.T) {
	err := New().
		Timeout(time.Millisecond * 100).
		Go(func() {
			time.Sleep(time.Second)
		})

	assert.Equal(t, ErrTimeout, err)
}

func TestWaitGo3(t *testing.T) {
	result := make(chan int, 10)

	err := New().
		Timeout(-1).
		Go(func() {
			result <- 1
			time.Sleep(time.Millisecond * 50)
		})

	assert.Nil(t, err)
	assert.Equal(t, 1, <-result)
}

func TestRecover(t *testing.T) {
	errStr := make(chan string, 10)

	err := New().
		EnsureStarted().
		Recover(func(e interface{}) {
			errStr <- fmt.Sprintf("%v", e)
		}).
		Go(func() {
			panic(`1`)
		})

	assert.Nil(t, err)
	assert.Equal(t, "1", <-errStr)
}

func TestAfter1(t *testing.T) {
	result := `1`

	err := New().
		EnsureStarted().
		After(func() {
			result += `3`
		}).
		Go(func() {
			result += `2`
		})

	assert.Nil(t, err)
	assert.Equal(t, "123", result)
}

func TestAfter2(t *testing.T) {
	result := `1`

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

	assert.Nil(t, err)
	assert.Equal(t, "123", result)
}

func TestAfter3(t *testing.T) {
	result := `1`

	err := New().
		EnsureStarted().
		Recover(func(e interface{}) {}).
		After(func() {
			result += `3`
		}).
		Go(func() {
			result += `2`
			panic(`ERR`)
		})

	assert.Nil(t, err)
	assert.Equal(t, "12", result)
}

func TestBefore1(t *testing.T) {
	result := `1`

	err := New().
		EnsureStarted().
		Before(func() {
			result = `0` + result
		}).
		Go(func() {
			result += `2`
		})

	assert.Nil(t, err)
	assert.Equal(t, "012", result)
}

func TestBefore2(t *testing.T) {
	result := `1`

	err := New().
		Recover(func(e interface{}) {}).
		EnsureStarted().
		Before(func() {
			result = `0` + result
		}).
		Go(func() {
			result += `2`
			panic(`XYZ`)
		})

	assert.Nil(t, err)
	assert.Equal(t, "012", result)
}

func TestAddToGroup(t *testing.T) {
	result := make(chan int, 10)
	wg := &sync.WaitGroup{}

	err1 := New().
		AddToGroup(wg).
		Go(func() {
			result <- 1
		})
	err2 := New().
		AddToGroup(wg).
		Go(func() {
			result <- 2
		})

	assert.Nil(t, err1)
	assert.Nil(t, err2)
	wg.Wait()
	close(result)
	sum := 0
	for v := range result {
		sum += v
	}
	assert.Equal(t, 3, sum)
}

func TestWithContext(t *testing.T) {
	result := make(chan int, 10000)
	ctx := context.Background()

	err := New().
		EnsureStarted().
		Timeout(time.Millisecond*110).
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

	assert.Equal(t, ErrTimeout, err)
	close(result)
	sum := 0
	for v := range result {
		sum += v
	}
	assert.Equal(t, 10, sum)
}

func TestWithContext2(t *testing.T) {
	result := make(chan int, 1000)
	ctx, cancel := context.WithCancel(context.Background())

	err := New().
		EnsureStarted().
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

	assert.Nil(t, err)

	<-time.After(time.Millisecond * 100)

	cancel()
	sum := 0
	for v := range result {
		sum += v
	}
	assert.Equal(t, 10, sum)
}

func TestWithContext3(t *testing.T) {
	result := make(chan int, 1000)
	ctx, cancel := context.WithCancel(context.Background())

	err := New().
		EnsureStarted().
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

	assert.Nil(t, err)

	<-time.After(time.Millisecond * 100)

	cancel()
	sum := 0
	for v := range result {
		sum += v
	}
	assert.Equal(t, 12, sum)
}

var (
	wg = &sync.WaitGroup{}
)

func simpleSupervisor(intensity int, fn func()) {
	if intensity <= 0 {
		return
	}
	intensity--
	New().
		AddToGroup(wg).
		Recover(func(e interface{}) {
			time.Sleep(time.Millisecond * 10)
			go simpleSupervisor(intensity, fn)
		}).
		Go(fn)
}

func TestSimpleSupervisor(t *testing.T) {
	cnt := new(int)
	*cnt = 0
	go simpleSupervisor(3, func() {
		*cnt = *cnt + 1
		panic(`1`)
	})

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond * 50)
	}()

	wg.Wait()
	assert.Equal(t, 3, *cnt)
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
