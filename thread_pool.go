package cptest

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shettyh/threadpool"
)

// Runnable represents a callable object. It allows representing the closure
// of a function as well-defined struct fields with the function in question
// having no closure (the Run() method).
type Runnable interface {
	Run()
}

// RunnableFunc adapts basic golang functions for the Runnable interface.
type RunnableFunc func()

// Run calls the contained function.
func (f RunnableFunc) Run() {
	f()
}

// WorkerPool abstracts the notion of a thread pool. It allows for
// interoperation with different thread pool libraries.
type WorkerPool interface {
	Execute(task Runnable) error
	WorkerCount() int
}

// ThreadPool is a generic implementation of WorkerPool.
type ThreadPool struct {
	threadCount int
	threadPool  *threadpool.ThreadPool
}

// NewThreadPool creates a thread pool with a definite limit to the number
// of concurrent operations permitted.
func NewThreadPool(count int) *ThreadPool {
	return &ThreadPool{
		threadCount: count,
		threadPool:  threadpool.NewThreadPool(count, int64(count)),
	}
}

// Execute tries to find a free worker and assign task to it. If no worker
// is available, returns ErrNoWorkerAvailable.
func (p *ThreadPool) Execute(task Runnable) error {
	return p.threadPool.Execute(threadpool.Runnable(task))
}

// WorkerCount returns the number of threads in a pool. These threads may
// may not have a task assigned to them.
func (p *ThreadPool) WorkerCount() int {
	return p.threadCount
}

// SpyThreadPool implements the same functionality as ThreadPool, but also
// tracks the goids of the threads involved in task execution. This is used
// to assert that an exact number of threads was invoked.
type SpyThreadPool struct {
	DirtyThreads map[int]struct{}
	mu           sync.Mutex

	threadCount int
	threadPool  *threadpool.ThreadPool
}

// NewSpyThreadPool creates a thread pool with the specified number of
// preallocated threads in the pool.
func NewSpyThreadPool(threadCount int) *SpyThreadPool {
	return &SpyThreadPool{
		DirtyThreads: make(map[int]struct{}),

		threadCount: threadCount,
		threadPool:  threadpool.NewThreadPool(threadCount, int64(threadCount)),
	}
}

// Borrowed from: https://gist.github.com/metafeather/3615b23097836bc36579100dac376906
func goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(string(buf[:n]))[1]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get gorountine id: %v", err))
	}

	return id
}

// Execute tries to find a worker to assign the task to. If it doesn't find
// one, it may or may NOT give an error. When error is returned is unspecified,
// but it shouldn't happen within general usage.
func (p *SpyThreadPool) Execute(task Runnable) error {
	return p.threadPool.Execute(RunnableFunc(func() {
		p.mu.Lock()
		p.DirtyThreads[goid()] = struct{}{}
		p.mu.Unlock()

		time.Sleep(5 * time.Millisecond)

		task.Run()
	}))
}

// WorkerCount returns the number of currently allocated threads. These may
// or may not have any tasks assigned to them.
func (p *SpyThreadPool) WorkerCount() int {
	return p.threadCount
}
