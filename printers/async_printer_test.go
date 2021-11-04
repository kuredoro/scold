package printers_test

import (
	"sync"
	"testing"
	"time"

	"github.com/kuredoro/scold"
	"github.com/kuredoro/scold/printers"
)

type blockingForwarder struct {
	stopBlocking chan struct{}
	receiver     scold.TestingEventListener
}

func newBlockingForwarder(receiver scold.TestingEventListener) *blockingForwarder {
	return &blockingForwarder{
		stopBlocking: make(chan struct{}),
		receiver:     receiver,
	}
}

func (f *blockingForwarder) TestStarted(id int) {
	select {
	case <-f.stopBlocking:
		f.receiver.TestStarted(id)
	}
}

func (f *blockingForwarder) TestFinished(test *scold.Test, result *scold.TestResult) {
	select {
	case <-f.stopBlocking:
		f.receiver.TestFinished(test, result)
	}

}

func (f *blockingForwarder) SuiteFinished(b *scold.TestingBatch) {
	select {
	case <-f.stopBlocking:
		f.receiver.SuiteFinished(b)
	}
}

func (f *blockingForwarder) StopBlocking() {
	close(f.stopBlocking)
}

func (f *blockingForwarder) StoppedBlocking() chan struct{} {
	return f.stopBlocking
}

func testResultWithID(id int) *scold.TestResult {
	result := &scold.TestResult{}
	result.ID = id
	return result
}

func TestAsyncEventForwarder(t *testing.T) {
	t.Run("is async", func(t *testing.T) {
		spy := &scold.SpyPrinter{}
		blockingF := newBlockingForwarder(spy)
		tests := []scold.Test{
			{Input: "1", Output: "1"},
		}

		asyncF := printers.NewAsyncEventForwarder(blockingF, 100)

		eventSent := make(chan struct{})

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			asyncF.TestStarted(1)
			asyncF.TestFinished(&tests[0], testResultWithID(1))
			asyncF.SuiteFinished(nil)

			close(eventSent)

			asyncF.Wait()
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			<-eventSent
			<-eventSent
			<-eventSent
			blockingF.StopBlocking()
			wg.Done()
		}()

		select {
		case <-time.After(time.Second):
			t.Error("the implementation is synchronous, want asynchronous")
			t.FailNow()
			return
		case <-blockingF.StoppedBlocking():
		}

		wg.Wait()

		scold.AssertListenerNotified(t, spy, tests)
	})
}
