package forwarders

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

// AsyncEventForwarder provides the ability to asynchronously process testing
// events. It can be used by users to make their own TestingEventListers
// asynchronous and prevent them from stalling the TestingBatch's event loop.
type AsyncEventForwarder struct {
	receiver    scold.TestingEventListener
	resultQueue chan testingEvent
	wg          util.WaitGroup
}

// NewAsyncEventForwarder will create an instance of AsyncEventForwarder that
// will forward events to the receiver. The internal queue will be of size
// queueSize. If the receiver cannot process events fast enough and queueSize
// is quite small, the AsyncEventForwarder will stall TestingBatch.
//
// Additional struct implementing util.WaitGroup can be also passed. It will be
// used instead of sync.WaitGroup.
func NewAsyncEventForwarder(receiver scold.TestingEventListener, queueSize int, args ...util.WaitGroup) *AsyncEventForwarder {
	var wg util.WaitGroup = &sync.WaitGroup{}

	if len(args) != 0 {
		wg = args[0]
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

// TestStarted pushes event to the internal queue and exits immediately.
func (f *AsyncEventForwarder) TestStarted(id int) {
	f.resultQueue <- testingEvent{eventType: testStartedEventType, id: id}
}

// TestFinished pushes event to the internal queue and exits immediately.
func (f *AsyncEventForwarder) TestFinished(test *scold.Test, result *scold.TestResult) {
	f.resultQueue <- testingEvent{eventType: testFinishedEventType, test: test, result: result}
}

// SuiteFinished pushes event to the internal queue, closes it and exits
// immediately.
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

// Wait will block until all events have been processed.
func (f *AsyncEventForwarder) Wait() {
	f.wg.Wait()
}
