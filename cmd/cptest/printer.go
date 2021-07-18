package main

import (
	"fmt"
	"time"

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
    cursor.ClearLinesUp(1)

	verdict := b.Verdicts[id]

	seconds := b.Times[id].Round(time.Millisecond).Seconds()
	fmt.Fprintf(stdout, "--- %s:\tTest %d (%.3fs)\n", verdictStr[verdict], id, seconds)

	if verdict != cptest.OK {
		fmt.Printf("Input:\n%s\n", res.test.Input)

		fmt.Fprintf(stdout, "Answer:\n%s\n", cptest.DumpLexemes(b.RichAnswers[id], diffColor))

		if verdict == cptest.RE {
			fmt.Printf("Exit code: %d\n\n", b.Outs[id].ExitCode)
			fmt.Printf("Output:\n%s\n", b.Outs[id].Stdout)
			fmt.Printf("Stderr:\n%s\n", b.Outs[id].Stderr)
		} else if verdict == cptest.WA {
			fmt.Fprintf(stdout, "Output:\n%s\n", cptest.DumpLexemes(b.RichOuts[id], diffColor))
			if b.Outs[id].Stderr != "" {
				fmt.Printf("Stderr:\n%s\n", b.Outs[id].Stderr)
			}
		} else if verdict == cptest.IE {
			fmt.Printf("Error:\n%v\n\n", b.Errs[id])
		}
	}

    fmt.Println()
    progressBar.Increment()
    progressBarRefresh <- struct{}{}
    time.Sleep(time.Millisecond)
}
