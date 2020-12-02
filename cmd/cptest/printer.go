package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/kuredoro/cptest"
	"github.com/logrusorgru/aurora"
)

const (
	ErrorColor = aurora.BoldFm | aurora.RedFg
)

var verdictStr = map[cptest.Verdict]aurora.Value{
	cptest.OK: aurora.Bold("OK").Green(),
	cptest.IE: aurora.Bold("IE"),
	cptest.WA: aurora.Bold("WA").Red(),
	cptest.RE: aurora.Bold("RE").Yellow(),
	cptest.TL: aurora.Bold("TL").Magenta(),
}

func RunPrinter(id int) {
	fmt.Printf("=== RUN\tTest %d\n", id)
}

func DumpLexemes(diffs []cptest.LexDiff) (string, string) {
	var gotStr, wantStr strings.Builder

	for _, diff := range diffs {
		if diff.Equal {
			gotStr.WriteString(diff.Got)
			if diff.Got != "\n" {
				gotStr.WriteRune(' ')
			}
			wantStr.WriteString(diff.Want)
			if diff.Want != "\n" {
				wantStr.WriteRune(' ')
			}
		} else {
			gotStr.WriteString(aurora.Colorize(diff.Got, ErrorColor).String())
			if diff.Got != "\n" {
				gotStr.WriteRune(' ')
			}
			wantStr.WriteString(aurora.Colorize(diff.Want, ErrorColor).String())
			if diff.Want != "\n" {
				wantStr.WriteRune(' ')
			}
		}
	}

	return gotStr.String(), wantStr.String()
}

func VerboseResultPrinter(b *cptest.TestingBatch, test cptest.Test, id int) {
	verdict := b.Verdicts[id]

	seconds := b.Times[id].Round(time.Millisecond).Seconds()
	fmt.Printf("--- %s:\tTest %d (%.3fs)\n", verdictStr[verdict], id, seconds)

	if verdict != cptest.OK {
		fmt.Printf("Input:\n%s\n", test.Input)

		output, answer := DumpLexemes(b.Diff)
		fmt.Printf("Answer:\n%s\n", answer)

		if verdict == cptest.RE {
			fmt.Printf("Stderr:\n%s\n", b.Outs[id])
		} else if verdict == cptest.WA {
			fmt.Printf("Output:\n%s\n", output)
		} else if verdict == cptest.IE {
			fmt.Printf("Error:\n%v\n\n", b.Errs[id])
		}
	}
}
