package cptest_test

import (
	"sync"
	"testing"

	"github.com/kuredoro/cptest"
)

func TestThreadPool(t *testing.T) {
    threadCount := 0
    var mu sync.Mutex

    pool := cptest.NewThreadPool(4)

    if pool.WorkerCount() != 4 {
        t.Errorf("got worker count = %d, want %d", pool.WorkerCount(), 4)
    }

    var barrier, wg sync.WaitGroup
    barrier.Add(4)

    for i := 0; i < 4; i++ {
        wg.Add(1)
        pool.Execute(cptest.RunnableFunc(func() {
            barrier.Done()
            barrier.Wait()

            mu.Lock()
            threadCount++
            mu.Unlock()

            wg.Done()
        }))
    }

    wg.Wait()

    if threadCount != 4 {
        t.Errorf("got %d different threads executed, want %d", threadCount, 4)
    }
}
