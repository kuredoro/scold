package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/atomicgo/cursor"
	"github.com/jonboulle/clockwork"
	"github.com/kuredoro/cptest"
	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-colorable"
)

var progressBar *ProgressBar

var stdout = colorable.NewColorableStdout()

type appArgs struct {
	Inputs     string   `arg:"-i" default:"inputs.txt" help:"file with tests"`
	NoColors   bool     `arg:"--no-colors" help:"disable colored output"`
    NoProgress bool     `arg:"--no-progress" help:"disable progress bar"`
    Jobs       int      `arg:"-j" placeholder:"COUNT" help:"Number of tests to run concurrently [default: CPU_COUNT]"`
	Executable string   `arg:"positional,required"`
    Args       []string `arg:"positional" placeholder:"ARG"`
}

var args appArgs

func (appArgs) Description() string {
	return `Feed programs fixed inputs, compare their outputs against expected ones.

Author: @kuredoro (github, twitter)
User manual: https://github.com/kuredoro/cptest
`
}

func (appArgs) Version() string {
	return "cptest 2.01a"
}

func mustParse(dest *appArgs) {
    parser, err := arg.NewParser(arg.Config{}, dest)
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: couldn't initialize command line argument parser")
        os.Exit(-1)
    }

    cliArgs := os.Args[1:]

    for end := 0; end != len(cliArgs)+1; end++ {

        fmt.Printf("parsing: %v\n", cliArgs[:end])

        // Skip flags until we find a bare string, the executable
        if end != 0 && cliArgs[end-1][0] == '-' {
            continue
        }

        err = parser.Parse(cliArgs[:end])
        if err != nil {
            fmt.Printf("error: %v\n", err)
            continue
        }

        fmt.Printf("Args: %#v\n", dest)

        fmt.Printf("Executable parsed, scraping arguments\n")

        dest.Args = cliArgs[end:]

        fmt.Printf("Final args: %#v\n", dest)
        break
    }

    // It will handle help and version arguments
    if err != nil {
        arg.MustParse(dest)
    }
}

func init() {
	mustParse(&args)

	if args.NoColors {
		cptest.Au = aurora.NewAurora(false)
	}

    if args.Jobs == 0 {
        args.Jobs = runtime.NumCPU()
    }

	verdictStr = map[cptest.Verdict]aurora.Value{
		cptest.OK: cptest.Au.Bold("OK").Green(),
		cptest.IE: cptest.Au.Bold("IE").Bold(),
		cptest.WA: cptest.Au.Bold("WA").Red(),
		cptest.RE: cptest.Au.Bold("RE").Magenta(),
		cptest.TL: cptest.Au.Bold("TL").Yellow(),
	}
}

func main() {
	inputsPath, err := filepath.Abs(args.Inputs)
	if err != nil {
		fmt.Printf("error: retreive inputs absolute path: %v\n", err)
		return
	}

	inputs, errs := readInputs(inputsPath)
	if errs != nil {
		for _, err := range errs {
			fmt.Printf("error: %v\n", err)
		}

		return
	}

	execPath, err := findFile(args.Executable)
	if err != nil {
		fmt.Printf("error: find executable: %v\n", err)
		return
	}

	proc := &Executable{
		Path: execPath,
		Args: args.Args,
	}

	TL := getTL(inputs)
    swatch := &cptest.ConfigurableStopwatcher{
        TL: TL,
        Clock: clockwork.NewRealClock(),
    }
    pool := cptest.NewThreadPool(args.Jobs)

	batch := cptest.NewTestingBatch(inputs, proc, swatch, pool)

    if TL == 0 {
        fmt.Println("time limit: infinity")
    } else {
        fmt.Printf("time limit: %v\n", TL)
    }
	fmt.Printf("floating point precision: %d digit(s)\n", batch.Lx.Precision)
    fmt.Printf("job count: %d\n", args.Jobs)

	batch.TestEndCallback = verboseResultPrinter

    var testingHeader string
    if args.NoColors {
        testingHeader = "    Testing"
    } else {
        testingHeader = aurora.Bold(aurora.Cyan("    Testing")).String()
    }

    progressBar = &ProgressBar{
        Total: len(inputs.Tests),
        Width: 20,
        Header: testingHeader,
    }

    if !args.NoProgress {
        fmt.Fprint(stdout, progressBar.String())
        cursor.StartOfLine()
    }

    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        verboseResultPrinterWorker()
        wg.Done()
    }()

	batch.Run()

    close(printQueue)
    wg.Wait()

    if !args.NoProgress {
        cursor.ClearLine()
        // wtf knows what's the behavior of the cursor packaage
        // Why it's outputting everything fine in verbose printer but here,
        // it clears the line but doesn't move the cursor???
        cursor.StartOfLine()
    }

	passCount := 0
	for _, v := range batch.Verdicts {
		if v == cptest.OK {
			passCount++
		}
	}

	if passCount == len(batch.Verdicts) {
		fmt.Fprintln(stdout, cptest.Au.Bold("OK").Green())
	} else {
		fmt.Fprintln(stdout, cptest.Au.Bold("FAIL").Red())
		fmt.Fprintf(stdout, "%d/%d passed\n", passCount, len(batch.Verdicts))
	}
}
