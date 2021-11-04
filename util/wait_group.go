package util

import (
    "sync"
)

type WaitGroup interface {
	Add(int)
	Done()
	Wait()
}

type SpyWaitGroup struct {
	mx sync.Mutex
	wg sync.WaitGroup

	Deltas        []int
	DeltaGoids    []int
	Awaited       bool
	AwaitedByGoid int
}

func (wg *SpyWaitGroup) Add(delta int) {
	wg.mx.Lock()
	wg.Deltas = append(wg.Deltas, delta)
    wg.DeltaGoids = append(wg.DeltaGoids, Goid())
	wg.mx.Unlock()

	wg.wg.Add(delta)
}

func (wg *SpyWaitGroup) Done() {
	wg.Add(-1)
}

func (wg *SpyWaitGroup) Wait() {
	wg.mx.Lock()
	wg.Awaited = true
    wg.AwaitedByGoid = Goid()
	wg.mx.Unlock()

	wg.wg.Wait()
}

