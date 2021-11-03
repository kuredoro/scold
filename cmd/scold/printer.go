package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/atomicgo/cursor"
	"github.com/kuredoro/scold"
	"github.com/logrusorgru/aurora"
)

const diffColor = aurora.RedFg
const missingNewlineColor = aurora.MagentaFg

// Initialized in init()
var verdictStr map[scold.Verdict]aurora.Value

// printQueue facilitates synchronized output of the test results, since End
// callback can be called simultaneously.
var printQueue = make(chan testResultNotification, 100)

type testResultNotification struct {
	test  *scold.Test
    result *scold.TestResult
}

func verboseResultPrinter(test *scold.Test, result *scold.TestResult) {
	printQueue <- testResultNotification{test, result}
}

func verboseResultPrinterWorker() {
	for result := range printQueue {
		printVerboseResult(result)
	}
}

func printAlwaysWithNewline(r io.Writer, text string) {
    fmt.Fprint(r, text)
    if text != "" && text[len(text)-1] != '\n' {
        fmt.Fprint(r, scold.DumpLexemes([]scold.RichText{{Str: "\n", Mask: []bool{true}}}, missingNewlineColor))
        fmt.Fprintln(r)
    }
}

func printVerboseResult(blob testResultNotification) {
	str := &strings.Builder{}

    test := blob.test
    result := blob.result

	verdict := result.Verdict

	seconds := result.Time.Round(time.Millisecond).Seconds()
	fmt.Fprintf(str, "--- %s:\tTest %d (%.3fs)\n", verdictStr[verdict], result.ID, seconds)

	if verdict != scold.OK {
		fmt.Fprintf(str, "Input:\n%s\n", test.Input)

		fmt.Fprintf(str, "Answer:\n%s\n", scold.DumpLexemes(result.RichAnswer, diffColor))

		if verdict == scold.RE {
			fmt.Fprintf(str, "Exit code: %d\n\n", result.Out.ExitCode)
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

	if !args.NoProgress {
		//progressBar.Current++

		//fmt.Fprint(str, progressBar.String())

		cursor.ClearLine()
	}

	fmt.Fprint(stdout, str)

	if !args.NoProgress {
		cursor.StartOfLine()
	}
}
