package main

import (
	"fmt"
	"time"
    "strings"

	"github.com/kuredoro/cptest"
	"github.com/logrusorgru/aurora"
    "github.com/atomicgo/cursor"
)

const diffColor = aurora.RedFg

// Initialized in init()
var verdictStr map[cptest.Verdict]aurora.Value

type TestResultNotification struct {
    batch *cptest.TestingBatch
    test cptest.Test
    id int
}

// printQueue facilitates synchronized output of the test results, since End
// callback can be called simultaneously.
var printQueue = make(chan *TestResultNotification, 100)

func verboseResultPrinter(b *cptest.TestingBatch, test cptest.Test, id int) {
    printQueue <- &TestResultNotification{b, test, id}
}

func verboseResultPrinterWorker() {
    for result := range printQueue {
        printVerboseResult(result)
    }
}

func printVerboseResult(res *TestResultNotification) {
    b := res.batch
    id := res.id

    str := &strings.Builder{}

	verdict := b.Verdicts[id]

	seconds := b.Times[id].Round(time.Millisecond).Seconds()
	fmt.Fprintf(str, "--- %s:\tTest %d (%.3fs)\n", verdictStr[verdict], id, seconds)

	if verdict != cptest.OK {
		fmt.Fprintf(str, "Input:\n%s\n", res.test.Input)

		fmt.Fprintf(str, "Answer:\n%s\n", cptest.DumpLexemes(b.RichAnswers[id], diffColor))

		if verdict == cptest.RE {
			fmt.Fprintf(str, "Exit code: %d\n\n", b.Outs[id].ExitCode)
			fmt.Fprintf(str, "Output:\n%s\n", b.Outs[id].Stdout)
			fmt.Fprintf(str, "Stderr:\n%s\n", b.Outs[id].Stderr)
		} else if verdict == cptest.WA {
			fmt.Fprintf(str, "Output:\n%s\n", cptest.DumpLexemes(b.RichOuts[id], diffColor))
			if b.Outs[id].Stderr != "" {
				fmt.Fprintf(str, "Stderr:\n%s\n", b.Outs[id].Stderr)
			}
		} else if verdict == cptest.IE {
			fmt.Fprintf(str, "Error:\n%v\n\n", b.Errs[id])
		}
	}

    progressBar.Current++

    fmt.Fprint(str, progressBar.String())

    cursor.ClearLine()
    fmt.Fprintf(stdout, str.String())
    cursor.StartOfLine()
}
