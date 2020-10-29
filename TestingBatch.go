package cptest

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// InternalError represents an error that occured due to internal failure in
// TestingBatch.
type InternalError bool

func (e InternalError) Error() string {
    return "internal"
}

const internalErr = InternalError(true)


// Verdict represents a verdict asssigned by the judge.
type Verdict int

// The set of all possible judge verdicts that can be assigned. The 
// abbreviatons are due to competitive programming online judges.
// (Except for IE, that stands for Internal Error)
const (
    OK Verdict = iota
    IE
    WA
    RE
    TL
)

var verdictStr = map[Verdict]string{
    OK: "OK",
    IE: "IE",
    WA: "WA",
    RE: "RE",
    TL: "TL",
}

// PrintResultFunc is type representing a function to print statistics and
// information about the finished test case.
type PrintResultFunc func(*TestingBatch, Test, int)

// BlankResultPrinter is the standard PrintResultFunc that outputs nothing.
func BlankResultPrinter(b *TestingBatch, test Test, id int) {}

func VerboseResultPrinter(b *TestingBatch, test Test, id int) {
    verdict := b.Stat[id]

    seconds := b.Times[id].Round(time.Millisecond).Seconds()
    fmt.Printf("--- %s:\tTest %d (%.3fs)\n", verdictStr[verdict], id, seconds)

    if verdict != OK {
        fmt.Printf("Input:\n%s\n\n", test.Input)
        fmt.Printf("Answer:\n%s\n\n", test.Output)

        if verdict == RE {
            fmt.Printf("Stderr:\n%s\n\n", b.outs[id])
        } else if verdict == WA {
            fmt.Printf("Output:\n%s\n\n", b.outs[id])
        } else if verdict == IE {
            fmt.Printf("Error:\n%v\n\n", b.errs[id])
        }
    }
}


// TestingBatch is responsible for running tests and evaluating the verdicts
// for tests. For each test case, the verdict and execution time are stored.
// It utilizer an instance of Processer to run tests, and an instance of
// Stopwatcher to track time limit. Optionally, user can set ResultPrinter
// to a custom function to output useful statistics about test case's result.
type TestingBatch struct {
    inputs Inputs

    complete chan int

    mu sync.Mutex
    errs map[int]error
    outs map[int]string

    Stat map[int]Verdict
    Times map[int]time.Duration

    Proc Processer
    Swatch Stopwatcher
    ResultPrinter PrintResultFunc
}

// NewTestingBatch will initialize channels and maps inside TestingBatch and
// will assign respective dependency injections.
func NewTestingBatch(inputs Inputs, proc Processer, swatch Stopwatcher) *TestingBatch {
    return &TestingBatch{
        inputs: inputs,

        complete: make(chan int),
        errs: make(map[int]error),
        outs: make(map[int]string),

        Stat: make(map[int]Verdict),
        Times: make(map[int]time.Duration),

        Proc: proc,
        Swatch: swatch,
        ResultPrinter: VerboseResultPrinter,
    }
}

func (b *TestingBatch) launchTest(id int, in string) {
    go func() {
        defer func() {
            if e := recover(); e != nil {
                b.mu.Lock()
                defer b.mu.Unlock()

                b.errs[id] = fmt.Errorf("%w: %v", internalErr, e)
                b.outs[id] = ""
            }

            b.complete<-id
        }()

        buf := &bytes.Buffer{}
        err := b.Proc.Run(strings.NewReader(in), buf)

        b.mu.Lock()
        defer b.mu.Unlock()

        b.errs[id] = err
        b.outs[id] = strings.TrimSpace(buf.String())
    }()
}

// Run will lauch test cases in parallel and then will wait for each test to
// finish or for the time to timelimit. When a test is finished the verdict
// and the time it took to execute are remembered. Additionally, ResultPrinter
// is called on the test case's statistics. When a time limit is reached,
// each not-yet-judged test is assigned TL verdict and the ResultPrinter is
// also called on each test.
func (b *TestingBatch) Run() {
    for i, test := range b.inputs.Tests {
        b.launchTest(i + 1, test.Input)

        fmt.Printf("=== RUN\tTest %d\n", i + 1)
    }

    for range b.inputs.Tests {
        var id int

        select {
        case id = <-b.complete:
        case tl := <-b.Swatch.TimeLimit():

            for id = range b.inputs.Tests {
                if _, finished := b.Stat[id + 1]; !finished {
                    b.Stat[id + 1] = TL
                    b.Times[id + 1] = tl

                    b.ResultPrinter(b, b.inputs.Tests[id], id + 1)
                }
            }

            return
        }

        test := b.inputs.Tests[id - 1]

        b.Times[id] = b.Swatch.Elapsed()

        // So I have these ugly ifs, cuz I want the result be printed as it
        // arrives. In the previous version I got printing defered and these
        // ifs were super dope and readable (thanks to continue statements).
        // But defer in a loop executes the printing at the loop's very end
        // which is unfortunate...
        // So, here I go. I somebody knows a way prettify this part, I would be
        // very glad!
        if err := b.errs[id]; err != nil {
            if errors.Is(err, internalErr) {
                b.Stat[id] = IE
            } else {
                b.Stat[id] = RE
            }
        } else if test.Output != b.outs[id] {
            b.Stat[id] = WA
        } else {
            b.Stat[id] = OK
        }

        b.ResultPrinter(b, test, id)
    }
}
