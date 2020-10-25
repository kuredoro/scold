package cptest

import (
    "errors"
    "fmt"
)

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

    fmt.Printf("--- %s: Test %d\n", verdictStr[verdict], id)
}


type TestingBatch struct {
    inputs Inputs
    proc Processer

    Stat map[int]Verdict

    ResultPrinter PrintResultFunc
}

func NewTestingBatch(inputs Inputs, proc Processer) *TestingBatch {
    return &TestingBatch{
        inputs: inputs,
        proc: proc,
        Stat: make(map[int]Verdict),
        ResultPrinter: VerboseResultPrinter,
    }
}

func (b *TestingBatch) Run() {
    for i, test := range b.inputs.Tests {
        b.proc.Run(i + 1, test.Input)

        fmt.Printf("=== RUN: Test %d\n", i + 1)
    }

    for _, test := range b.inputs.Tests {
        id := b.proc.WaitCompleted()
        defer b.ResultPrinter(b, test, id)

        if err := b.proc.GetError(id); err != nil {
            if errors.Is(err, internalErr) {
                b.Stat[id] = IE
                continue
            }

            b.Stat[id] = RE
            continue
        }

        if test.Output != b.proc.GetOutput(id) {
            b.Stat[id] = WA
            continue
        }

        b.Stat[id] = OK
    }
}
