// Package goroutines provides utilities to perform common tasks on goroutines
package goroutines

import (
	"sync"
	"time"
)

// New creates an instange of Go struct
func New() Go { return Go{} }

// Go provides a fluent way to prepare & start a goroutine
type Go struct {
	waitStart   bool
	timeout     time.Duration
	recoverFunc func(interface{})
	before      func()
	after       func()
	deferAfter  bool
	wg          *sync.WaitGroup
}

// Go is final call in the fluent chain
func (x Go) Go(f func()) error {
	var started, funcDone chan struct{}
	if x.waitStart {
		started = make(chan struct{})
	}
	if x.timeout != 0 {
		funcDone = make(chan struct{})
	}
	if x.wg != nil {
		x.wg.Add(1)
	}

	go func() {
		if started != nil {
			close(started)
		}
		if x.wg != nil {
			defer x.wg.Done()
		}
		if funcDone != nil {
			defer close(funcDone)
		}
		if x.recoverFunc != nil {
			defer func() {
				if e := recover(); e != nil {
					x.recoverFunc(e)
				}
			}()
		}
		if x.before != nil {
			x.before()
		}
		if x.after != nil && x.deferAfter {
			defer x.after()
		}

		f()

		if x.after != nil && !x.deferAfter {
			x.after()
		}
	}()

	if started != nil {
		<-started
	}
	if funcDone != nil {
		if x.timeout > 0 {
			tm := time.NewTimer(x.timeout)
			defer func() {
				if !tm.Stop() {
					select {
					case <-tm.C:
					default:
					}
				}
			}()
			select {
			case <-funcDone:
			case <-tm.C:
				return ErrTimeout
			}
		} else if x.timeout < 0 {
			<-funcDone
		}
	}

	return nil
}

// WaitStart instructs Go to start a goroutine and wait for it to start,
// and after goroutine started, it returns.
func (x Go) WaitStart() Go {
	x.waitStart = true
	return x
}

// WaitGo waits for f to complete (in a goroutine)
// or times out (returning ErrTimeout). A negative value for timeout, means
// waiting infinitely.
func (x Go) WaitGo(timeout time.Duration) Go {
	x.timeout = timeout
	return x
}

// WaitGroup registers the goroutine in a sync.WaitGroup by adding necessary code (Add/Done)
func (x Go) WaitGroup(wg *sync.WaitGroup) Go {
	x.wg = wg
	return x
}

// Recover recovers from panic and returns error,
// or returns the provided error
func (x Go) Recover(recoverFunc func(interface{})) Go {
	x.recoverFunc = recoverFunc
	return x
}

// Before will be called before the goroutine func at the begining of the same goroutine
func (x Go) Before(before func()) Go {
	x.before = before
	return x
}

// After will get called after the goroutine func, it can be deferred
func (x Go) After(after func(), deferred ...bool) Go {
	x.after = after
	if len(deferred) > 0 {
		x.deferAfter = deferred[0]
	}
	return x
}

// Constants
const (
	ErrTimeout = _error(`TIMEOUT`)
)

type _error string

func (v _error) Error() string { return string(v) }
