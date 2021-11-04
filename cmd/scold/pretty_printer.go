package printers

import (
	"fmt"
	"sync"

	"github.com/kuredoro/scold"
)

const prettyPrinterQueueSize = 100

type testFinishedEvent struct {
	test  *scold.Test
    result *scold.TestResult
}

type PrettyPrinter struct {
    resultQueue chan testFinishedEvent
    asyncPrinterStarted bool
    wg sync.WaitGroup
}

func NewPrettyPrinter() *PrettyPrinter {
    return &PrettyPrinter{
        resultQueue: make(chan testFinishedEvent, prettyPrinterQueueSize),
    }
}

func (p *PrettyPrinter) TestStarted(id int) {
    if !p.asyncPrinterStarted {
        p.wg.Add(1)
        p.asyncPrinterStarted = true

        go p.asyncPrinter()
    }
}

func (p *PrettyPrinter) TestFinished(test *scold.Test, result *scold.TestResult) {
	p.resultQueue <- testFinishedEvent{test, result}
}

func (p *PrettyPrinter) SuiteFinished(b *scold.TestingBatch) {
    close(p.resultQueue)
    p.wg.Wait()
    fmt.Printf("Suite finished")
}

func (p *PrettyPrinter) asyncPrinter() {
    for event := range p.resultQueue {
        //test := event.test
        result := event.result
        fmt.Printf("Test %d finished\n", result.ID)
    }

    p.wg.Done()
}
