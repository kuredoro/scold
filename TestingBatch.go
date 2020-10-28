package cptest

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
)

type InternalError bool

func (e InternalError) Error() string {
    return "internal"
}

const internalErr = InternalError(true)


type Verdict int

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

type PrintResultFunc func(*TestingBatch, Test, int)

func BlankResultPrinter(b *TestingBatch, test Test, id int) {}

func VerboseResultPrinter(b *TestingBatch, test Test, id int) {
    verdict := b.Stat[id]

    fmt.Printf("--- %s:\tTest %d\n", verdictStr[verdict], id)

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



type TestingBatch struct {
    inputs Inputs

    complete chan int

    mu sync.Mutex
    errs map[int]error
    outs map[int]string

    Stat map[int]Verdict

    Proc Processer
    Swatch Stopwatcher
    ResultPrinter PrintResultFunc
}

func NewTestingBatch(inputs Inputs, proc Processer, swatch Stopwatcher) *TestingBatch {
    return &TestingBatch{
        inputs: inputs,

        complete: make(chan int),
        errs: make(map[int]error),
        outs: make(map[int]string),

        Stat: make(map[int]Verdict),

        Proc: proc,
        Swatch: swatch,
        ResultPrinter: VerboseResultPrinter,
    }
}

func (b *TestingBatch) launchTest(id int, in string) {
    go func() {
        defer func() {
            b.mu.Lock()
            defer b.mu.Unlock()

            if e := recover(); e != nil {
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

func (b *TestingBatch) Run() {
    for i, test := range b.inputs.Tests {
        b.launchTest(i + 1, test.Input)

        fmt.Printf("=== RUN\tTest %d\n", i + 1)
    }

    for range b.inputs.Tests {
        var id int

        select {
        case id = <-b.complete:
        case <-b.Swatch.TimeLimit():

            for id = range b.inputs.Tests {
                if _, finished := b.Stat[id + 1]; !finished {
                    b.Stat[id + 1] = TL
                    b.ResultPrinter(b, b.inputs.Tests[id], id + 1)
                }
            }

            return
        }

        test := b.inputs.Tests[id - 1]

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
