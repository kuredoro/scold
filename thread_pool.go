package cptest

import (
	"runtime"
	"strconv"
	"strings"
	"time"
    "fmt"
    "sync"

	"github.com/shettyh/threadpool"
)

// Runnable is an interface that represents a callable object.
type Runnable interface {
    Run()
}

type RunnableFunc func()

func (f RunnableFunc) Run() {
    f()
}

// WorkerPoll is an interface that abstracts the notion of a thread pool.
type WorkerPool interface {
    Execute(task Runnable) error
}

type ThreadPool struct {
    threadPool *threadpool.ThreadPool
}

func NewThreadPool(count int) *ThreadPool {
    return &ThreadPool{
        threadPool: threadpool.NewThreadPool(count, int64(count)),
    }
}

func (p *ThreadPool) Execute(task Runnable) error {
    return p.threadPool.Execute(threadpool.Runnable(task))
}

type SpyThreadPool struct {
    DirtyThreads map[int]struct{}
    mu sync.Mutex

    threadPool *threadpool.ThreadPool
}

func NewSpyThreadPool(threadCount int) *SpyThreadPool {
    return &SpyThreadPool{
        DirtyThreads: make(map[int]struct{}),

        threadPool: threadpool.NewThreadPool(threadCount, int64(threadCount)),
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

func (p *SpyThreadPool) Execute(task Runnable) error {
    return p.threadPool.Execute(RunnableFunc(func() {
        p.mu.Lock()
        p.DirtyThreads[goid()] = struct{}{}
        p.mu.Unlock()

        time.Sleep(5 * time.Millisecond)
        
        task.Run()
    }))
}


