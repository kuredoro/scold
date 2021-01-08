package main

import (
	"fmt"
	"time"

	"github.com/kuredoro/cptest"
	"github.com/logrusorgru/aurora"
)

const (
	DiffColor = aurora.RedFg
)

// Initialized in init()
var verdictStr map[cptest.Verdict]aurora.Value

func RunPrinter(id int) {
	fmt.Printf("=== RUN\tTest %d\n", id)
}

func VerboseResultPrinter(b *cptest.TestingBatch, test cptest.Test, id int) {
	verdict := b.Verdicts[id]

	seconds := b.Times[id].Round(time.Millisecond).Seconds()
	fmt.Fprintf(stdout, "--- %s:\tTest %d (%.3fs)\n", verdictStr[verdict], id, seconds)

	if verdict != cptest.OK {
		fmt.Printf("Input:\n%s\n", test.Input)

		fmt.Fprintf(stdout, "Answer:\n%s\n", cptest.DumpLexemes(b.RichAnswers[id], DiffColor))

		if verdict == cptest.RE {
			fmt.Printf("Exit code: %d\n\n", b.Outs[id].ExitCode)
			fmt.Printf("Output:\n%s\n", b.Outs[id].Stdout)
			fmt.Printf("Stderr:\n%s\n", b.Outs[id].Stderr)
		} else if verdict == cptest.WA {
			fmt.Fprintf(stdout, "Output:\n%s\n", cptest.DumpLexemes(b.RichOuts[id], DiffColor))
            if b.Outs[id].Stderr != "" {
                fmt.Printf("Stderr:\n%s\n", b.Outs[id].Stderr)
            }
		} else if verdict == cptest.IE {
			fmt.Printf("Error:\n%v\n\n", b.Errs[id])
		}
	}
}
