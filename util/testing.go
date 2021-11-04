package util

import (
	"testing"
)

func AssertSpyWaitGroupNormalUsage(t *testing.T, wg *SpyWaitGroup, threadCount int) {
	t.Helper()

	if !wg.Awaited {
		t.Error("no Wait() was called on the WaitGroup, want one")
	}

	if len(wg.Deltas) == 0 {
		t.Log("no Add() or Done() was called on the WaitGroup")
		return
	}

	balance := 0
	for _, delta := range wg.Deltas {
		balance += delta
	}

	if balance > 0 {
		t.Error("not all goroutine called Done() on WaitGroup, want all")
	} else if balance < 0 {
		t.Error("more Done() than Add() was called on the WaitGroup, want the same number")
	}

	// Add should be called from the same goroutine.
	addCaller := wg.DeltaGoids[0]
	for i, delta := range wg.Deltas {
		if delta < 0 {
			continue
		}

		if addCaller != wg.DeltaGoids[i] {
			t.Errorf("WaitGroup Add() call #%d was performed from a different goroutine, want the same.", i+1)
		}
	}

	doneCallers := map[int]struct{}{}
	for i, delta := range wg.Deltas {
		if delta > 0 {
			continue
		}

		if delta < -1 {
			t.Errorf("got a call to Add() with value %d, want none", delta)
			continue
		}

		_, exists := doneCallers[wg.DeltaGoids[i]]
		if exists {
			t.Errorf("the goroutine #%d called Done() several times, want once", wg.DeltaGoids[i])
		}

		doneCallers[wg.DeltaGoids[i]] = struct{}{}
	}

	_, doneCalledByMain := doneCallers[addCaller]
	if doneCalledByMain {
		t.Errorf("Waitgroup Add() and Done() were called from the same goroutine, want different")
	}

	if addCaller != wg.AwaitedByGoid {
		t.Errorf("WaitGroup Add() and Wait() are not called from the same goroutine, want the same")
	}
}
