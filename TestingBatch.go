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
)

var verdictStr = map[Verdict]string{
    OK: "OK",
    IE: "IE",
    WA: "WA",
    RE: "RE",
}

type PrintResultFunc func(*TestingBatch, Test, int)

func BlankResultPrinter(b *TestingBatch, test Test, id int) {}

func VerboseResultPrinter(b *TestingBatch, test Test, id int) {
    verdict := b.Stat[id]

    fmt.Printf("--- %s:\tTest %d\n", verdictStr[verdict], id)

    if verdict != OK {
        fmt.Printf("Test:\n%s\n\n", test.Input)
        fmt.Printf("Answer:\n%s\n\n", test.Output)

        if verdict == RE {
            fmt.Printf("Stderr:\n%s\n\n", b.outs[id])
        } else if verdict == WA {
            fmt.Printf("Program's output:\n%s\n\n", b.outs[id])
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
    ResultPrinter PrintResultFunc
}

func NewTestingBatch(inputs Inputs, proc Processer) *TestingBatch {
    return &TestingBatch{
        inputs: inputs,

        complete: make(chan int),
        errs: make(map[int]error),
        outs: make(map[int]string),

        Stat: make(map[int]Verdict),
        Proc: proc,
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
        id := <-b.complete

        test := b.inputs.Tests[id - 1]
        defer b.ResultPrinter(b, test, id)

        if err := b.errs[id]; err != nil {
            if errors.Is(err, internalErr) {
                b.Stat[id] = IE
                continue
            }

            b.Stat[id] = RE
            continue
        }

        if test.Output != b.outs[id] {
            b.Stat[id] = WA
            continue
        }

        b.Stat[id] = OK
    }
}
