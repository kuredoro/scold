package main

import (
	"fmt"
	"time"

	"github.com/kuredoro/cptest"
)

var verdictStr = map[cptest.Verdict]string{
	cptest.OK: "OK",
	cptest.IE: "IE",
	cptest.WA: "WA",
	cptest.RE: "RE",
	cptest.TL: "TL",
}

func RunPrinter(id int) {
    fmt.Printf("=== RUN\tTest %d\n", id)
}

func VerboseResultPrinter(b *cptest.TestingBatch, test cptest.Test, id int) {
	verdict := b.Stat[id]

	seconds := b.Times[id].Round(time.Millisecond).Seconds()
	fmt.Printf("--- %s:\tTest %d (%.3fs)\n", verdictStr[verdict], id, seconds)

	if verdict != cptest.OK {
		fmt.Printf("Input:\n%s\n", test.Input)
		fmt.Printf("Answer:\n%s\n", test.Output)

		if verdict == cptest.RE {
			fmt.Printf("Stderr:\n%s\n", b.Outs[id])
		} else if verdict == cptest.WA {
			fmt.Printf("Output:\n%s\n", b.Outs[id])
		} else if verdict == cptest.IE {
			fmt.Printf("Error:\n%v\n\n", b.Errs[id])
		}
	}
}
