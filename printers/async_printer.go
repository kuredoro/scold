package printers

import (
	"sync"

	"github.com/kuredoro/scold"
	"github.com/kuredoro/scold/util"
)

const (
	testStartedEventType = iota
	testFinishedEventType
	suiteFinishedEventType
)

type testingEvent struct {
	eventType int
	id        int
	test      *scold.Test
	result    *scold.TestResult
	b         *scold.TestingBatch
}

func (e *testingEvent) testStarted() int {
	return e.id
}

func (e *testingEvent) testFinished() (*scold.Test, *scold.TestResult) {
	return e.test, e.result
}

func (e *testingEvent) suiteFinished() *scold.TestingBatch {
	return e.b
}

type Latch interface {
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
    wg.DeltaGoids = append(wg.DeltaGoids, util.Goid())
	wg.mx.Unlock()

	wg.wg.Add(delta)
}

func (wg *SpyWaitGroup) Done() {
	wg.Add(-1)
}

func (wg *SpyWaitGroup) Wait() {
	wg.mx.Lock()
	wg.Awaited = true
    wg.AwaitedByGoid = util.Goid()
	wg.mx.Unlock()

	wg.wg.Wait()
}

type AsyncEventForwarder struct {
	receiver    scold.TestingEventListener
	resultQueue chan testingEvent
	wg          Latch
}

func NewAsyncEventForwarder(receiver scold.TestingEventListener, queueSize int, args ...Latch) *AsyncEventForwarder {
	var wg Latch = &sync.WaitGroup{}

	if len(args) != 0 {
		wg = args[0].(Latch)
	}

	asyncF := &AsyncEventForwarder{
		receiver:    receiver,
		resultQueue: make(chan testingEvent, queueSize),
		wg:          wg,
	}

	asyncF.wg.Add(1)
	go asyncF.asyncPrinter()

	return asyncF
}

func (f *AsyncEventForwarder) TestStarted(id int) {
	f.resultQueue <- testingEvent{eventType: testStartedEventType, id: id}
}

func (f *AsyncEventForwarder) TestFinished(test *scold.Test, result *scold.TestResult) {
	f.resultQueue <- testingEvent{eventType: testFinishedEventType, test: test, result: result}
}

func (f *AsyncEventForwarder) SuiteFinished(b *scold.TestingBatch) {
	f.resultQueue <- testingEvent{eventType: suiteFinishedEventType, b: b}

	close(f.resultQueue)
}

func (f *AsyncEventForwarder) asyncPrinter() {
	for event := range f.resultQueue {
		switch event.eventType {
		case testStartedEventType:
			f.receiver.TestStarted(event.testStarted())
		case testFinishedEventType:
			f.receiver.TestFinished(event.testFinished())
		case suiteFinishedEventType:
			f.receiver.SuiteFinished(event.suiteFinished())
		}
	}

	f.wg.Done()
}

func (f *AsyncEventForwarder) Wait() {
	f.wg.Wait()
}
