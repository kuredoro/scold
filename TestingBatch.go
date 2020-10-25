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

    fmt.Printf("--- %s:\tTest %d\n", verdictStr[verdict], id)

    if verdict != OK {
        fmt.Printf("Test:\n%s\n\n", test.Input)
        fmt.Printf("Answer:\n%s\n\n", test.Output)

        switch verdict {
        case RE:
            fmt.Println("Stderr:")
        case WA:
            fmt.Println("Program's output:")
        }

        fmt.Printf("%s\n\n", b.proc.GetOutput(id))
    }
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

        fmt.Printf("=== RUN\tTest %d\n", i + 1)
    }

    for range b.inputs.Tests {
        id := b.proc.WaitCompleted()
        test := b.inputs.Tests[id - 1]
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
