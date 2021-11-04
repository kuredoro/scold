package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/atomicgo/cursor"
	"github.com/kuredoro/scold"
	"github.com/kuredoro/scold/util"
	"github.com/logrusorgru/aurora"
)

const diffColor = aurora.RedFg
const missingNewlineColor = aurora.MagentaFg

type PrettyPrinter struct {
	Bar        *ProgressBar
	verdictStr map[scold.Verdict]aurora.Value
}

func NewPrettyPrinter(au aurora.Aurora) *PrettyPrinter {
	p := &PrettyPrinter{}

	p.verdictStr = map[scold.Verdict]aurora.Value{
		scold.OK: au.Bold("OK").Green(),
		scold.IE: au.Bold("IE").Bold(),
		scold.WA: au.Bold("WA").BrightRed(),
		scold.RE: au.Bold("RE").Magenta(),
		scold.TL: au.Bold("TL").Yellow(),
	}

	return p
}

func (p *PrettyPrinter) TestStarted(id int) {
	if p.Bar != nil {
		fmt.Fprint(stdout, p.Bar.String())
		cursor.StartOfLine()
	}
}

func (p *PrettyPrinter) TestFinished(test *scold.Test, result *scold.TestResult) {
	str := &strings.Builder{}

	verdict := result.Verdict

	seconds := result.Time.Round(time.Millisecond).Seconds()
	fmt.Fprintf(str, "--- %s:\tTest %d (%.3fs)\n", p.verdictStr[verdict], result.ID, seconds)

	if verdict != scold.OK {
		fmt.Fprintf(str, "Input:\n%s\n", test.Input)

		fmt.Fprintf(str, "Answer:\n%s\n", scold.DumpLexemes(result.RichAnswer, diffColor))

		if verdict == scold.RE {
            if util.IsPossiblyNegative(result.Out.ExitCode) {
                fmt.Fprintf(str, "Exit code: %d (unsigned: %d)\n\n", int32(result.Out.ExitCode),
                    uint64(result.Out.ExitCode))
            } else {
                fmt.Fprintf(str, "Exit code: %d\n\n", result.Out.ExitCode)
            }
			fmt.Fprint(str, "Output:\n")
			printAlwaysWithNewline(str, result.Out.Stdout)
			fmt.Fprint(str, "Stderr:\n")
			printAlwaysWithNewline(str, result.Out.Stderr)
		} else if verdict == scold.WA {
			fmt.Fprintf(str, "Output:\n%s\n", scold.DumpLexemes(result.RichOut, diffColor))
			if result.Out.Stderr != "" {
				fmt.Fprintf(str, "Stderr:\n%s\n", result.Out.Stderr)
			}
		} else if verdict == scold.TL {
			if result.Out.Stdout != "" {
				fmt.Fprint(str, "Output:\n")
				printAlwaysWithNewline(str, result.Out.Stdout)
			}

			if result.Out.Stderr != "" {
				fmt.Fprint(str, "Stderr:\n")
				printAlwaysWithNewline(str, result.Out.Stderr)
			}
		} else if verdict == scold.IE {
			fmt.Fprintf(str, "Error:\n%v\n\n", result.Err)
		}
	}

	if p.Bar != nil {
		p.Bar.Current++

		fmt.Fprint(str, p.Bar.String())

		cursor.ClearLine()
	}

	fmt.Fprint(stdout, str)

	if p.Bar != nil {
		cursor.StartOfLine()
	}
}

func (p *PrettyPrinter) SuiteFinished(b *scold.TestingBatch) {
	if p.Bar != nil {
		cursor.ClearLine()
		// wtf knows what's the behavior of the cursor packaage
		// Why it's outputting everything fine in TestFinished but here,
		// it clears the line but doesn't move the cursor???
		cursor.StartOfLine()
	}

	passCount := 0
	for _, r := range b.Results {
		if r.Verdict == scold.OK {
			passCount++
		}
	}

	if passCount == len(b.Results) {
		fmt.Fprintln(stdout, scold.Au.Bold("OK").Green())
	} else {
		fmt.Fprintln(stdout, scold.Au.Bold("FAIL").Red())
		fmt.Fprintf(stdout, "%d/%d passed\n", passCount, len(b.Results))
	}
}

func printAlwaysWithNewline(r io.Writer, text string) {
	fmt.Fprint(r, text)
	if text != "" && text[len(text)-1] != '\n' {
		fmt.Fprint(r, scold.DumpLexemes([]scold.RichText{{Str: "\n", Mask: []bool{true}}}, missingNewlineColor))
		fmt.Fprintln(r)
	}
}
