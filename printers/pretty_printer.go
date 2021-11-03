package printers

import (
	"fmt"

	"github.com/kuredoro/scold"
)

type PrettyPrinter struct {

}

func (p *PrettyPrinter) TestStarted(id int) {
    fmt.Printf("Test %d started\n", id)
}

func (p *PrettyPrinter) TestFinished(test *scold.Test, result *scold.TestResult) {
    fmt.Printf("Test %d finished\n", result.ID)
}

func (p *PrettyPrinter) SuiteFinished(b *scold.TestingBatch) {
    fmt.Printf("Suite finished")
}
