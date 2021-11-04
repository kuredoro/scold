package printers

import (
	"sync"

	"github.com/kuredoro/scold"
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

type AsyncEventForwarder struct {
	receiver    scold.TestingEventListener
	resultQueue chan testingEvent
	wg          sync.WaitGroup
}

func NewAsyncEventForwarder(receiver scold.TestingEventListener, queueSize int) *AsyncEventForwarder {
	asyncF := &AsyncEventForwarder{
		receiver:    receiver,
		resultQueue: make(chan testingEvent, queueSize),
	}

	asyncF.wg.Add(1)
	go asyncF.asyncPrinter()

	return asyncF
}

func (p *AsyncEventForwarder) TestStarted(id int) {
	p.resultQueue <- testingEvent{eventType: testStartedEventType, id: id}
}

func (p *AsyncEventForwarder) TestFinished(test *scold.Test, result *scold.TestResult) {
	p.resultQueue <- testingEvent{eventType: testFinishedEventType, test: test, result: result}
}

func (p *AsyncEventForwarder) SuiteFinished(b *scold.TestingBatch) {
	p.resultQueue <- testingEvent{eventType: suiteFinishedEventType, b: b}

	close(p.resultQueue)
}

func (p *AsyncEventForwarder) asyncPrinter() {
	for event := range p.resultQueue {
		switch event.eventType {
		case testStartedEventType:
			p.receiver.TestStarted(event.testStarted())
		case testFinishedEventType:
			p.receiver.TestFinished(event.testFinished())
		case suiteFinishedEventType:
			p.receiver.SuiteFinished(event.suiteFinished())
		}
	}

	p.wg.Done()
}

func (p *AsyncEventForwarder) Wait() {
	p.wg.Wait()
}

/*
type AsyncEventForwarder struct {
    receiver scold.TestingEventListener
    resultQueue chan testingEvent
    asyncPrinterStarted bool
    wg sync.WaitGroup
}

func NewAsyncEventForwarder(receiver scold.TestingEventListener, queueSize int) *AsyncEventForwarder {
    return &AsyncEventForwarder{
        receiver: receiver,
        resultQueue: make(chan testingEvent, queueSize),
    }
}

func (p *AsyncEventForwarder) TestStarted(id int) {
    if !p.asyncPrinterStarted {
        p.wg.Add(1)
        p.asyncPrinterStarted = true

        go p.asyncPrinter()
    }

    p.resultQueue <- testingEvent{eventType: testStartedEventType, id: id}
}

func (p *AsyncEventForwarder) TestFinished(test *scold.Test, result *scold.TestResult) {
    p.resultQueue <- testingEvent{eventType: testFinishedEventType, test: test, result: result}
}

func (p *AsyncEventForwarder) SuiteFinished(b *scold.TestingBatch) {
    p.resultQueue <- testingEvent{eventType: suiteFinishedEventType, b: b}

    close(p.resultQueue)
    p.wg.Wait()
}

func (p *AsyncEventForwarder) asyncPrinter() {
    for event := range p.resultQueue {
        switch event.eventType {
        case testStartedEventType:
            p.receiver.TestStarted(event.testStarted())
        case testFinishedEventType:
            p.receiver.TestFinished(event.testFinished())
        case suiteFinishedEventType:
            p.receiver.SuiteFinished(event.suiteFinished())
        }
    }

    p.wg.Done()
}
*/
