package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/kuredoro/cptest"
	"github.com/logrusorgru/aurora"
)

const (
	DiffColor = aurora.ReverseFm
)

var verdictStr = map[cptest.Verdict]aurora.Value{
	cptest.OK: aurora.Bold("OK").Green(),
	cptest.IE: aurora.Bold("IE"),
	cptest.WA: aurora.Bold("WA").Red(),
	cptest.RE: aurora.Bold("RE").Red(),
	cptest.TL: aurora.Bold("TL").Yellow(),
}

func RunPrinter(id int) {
	fmt.Printf("=== RUN\tTest %d\n", id)
}

func DumpLexemes(xms []cptest.RichText) string {
	var str strings.Builder

	for _, xm := range xms {
        if xm.Str == "\n" {
            str.WriteRune('\n')
            continue
        }

        str.WriteString(xm.Colorize(DiffColor))
	}

	return str.String()
}

func VerboseResultPrinter(b *cptest.TestingBatch, test cptest.Test, id int) {
	verdict := b.Verdicts[id]

	seconds := b.Times[id].Round(time.Millisecond).Seconds()
	fmt.Printf("--- %s:\tTest %d (%.3fs)\n", verdictStr[verdict], id, seconds)

	if verdict != cptest.OK {
		fmt.Printf("Input:\n%s\n", test.Input)

		fmt.Printf("Answer:\n%s\n", DumpLexemes(b.Diff.Want))

		if verdict == cptest.RE {
			fmt.Printf("Stderr:\n%s\n", b.Outs[id])
		} else if verdict == cptest.WA {
			fmt.Printf("Output:\n%s\n", DumpLexemes(b.Diff.Got))
		} else if verdict == cptest.IE {
			fmt.Printf("Error:\n%v\n\n", b.Errs[id])
		}
	}
}
