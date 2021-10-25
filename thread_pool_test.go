package scold_test

import (
	"sync"
	"testing"

	"github.com/kuredoro/scold"
	"github.com/maxatome/go-testdeep/td"
)

func TestThreadPool(t *testing.T) {
	threadCount := 0
	var mu sync.Mutex

	pool := scold.NewThreadPool(4)

	if pool.WorkerCount() != 4 {
		t.Errorf("got worker count = %d, want %d", pool.WorkerCount(), 4)
	}

	var barrier, wg sync.WaitGroup
	barrier.Add(4)

	for i := 0; i < 4; i++ {
		wg.Add(1)
		err := pool.Execute(scold.RunnableFunc(func() {
			barrier.Done()
			barrier.Wait()

			mu.Lock()
			threadCount++
			mu.Unlock()

			wg.Done()
		}))

		td.CmpNoError(t, err)
	}

	wg.Wait()

	if threadCount != 4 {
		t.Errorf("got %d different threads executed, want %d", threadCount, 4)
	}
}
