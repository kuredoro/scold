package util

import (
	"sync"
)

// WaitGroup allows its users to inject custom implementations that conform
// to sync.WaitGroup API.
type WaitGroup interface {
	Add(int)
	Done()
	Wait()
}

// SpyWaitGroup provides an ability to assess whether the usage of the
// WaitGroup was correct. It records the sequence of calls as an array
// of deltas, arguments to the WaitGroup.Add() function. Additionally,
// it records the goid of the goroutines that performed the calls.
// SpyWaitGroup uses sync.WaitGroup under the hood and is thread-safe.
type SpyWaitGroup struct {
	mx sync.Mutex
	wg sync.WaitGroup

	Deltas        []int
	DeltaGoids    []int
	Awaited       bool
	AwaitedByGoid int
}

// Add remembers the delta and the goid of caller and forwards the
// call to sync.WaitGroup.
func (wg *SpyWaitGroup) Add(delta int) {
	wg.mx.Lock()
	wg.Deltas = append(wg.Deltas, delta)
	wg.DeltaGoids = append(wg.DeltaGoids, Goid())
	wg.mx.Unlock()

	wg.wg.Add(delta)
}

// Done is equivalent to Add(-1)
func (wg *SpyWaitGroup) Done() {
	wg.Add(-1)
}

// Wait remembers the goid of the caller and sets Awaited to true.
// It consequently, calls sync.WaitGroup.Wait().
func (wg *SpyWaitGroup) Wait() {
	wg.mx.Lock()
	wg.Awaited = true
	wg.AwaitedByGoid = Goid()
	wg.mx.Unlock()

	wg.wg.Wait()
}
