package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/atomicgo/cursor"
	"github.com/jonboulle/clockwork"
	"github.com/kuredoro/cptest"
	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-colorable"
)

var progressBar *ProgressBar

var stdout = colorable.NewColorableStdout()

type JobCount int

func (jc *JobCount) UnmarshalText(b []byte) error {
	if bytes.Equal(b, []byte("CPU_COUNT")) {
        *jc = JobCount(runtime.NumCPU())
		return nil
	}

	val, err := strconv.ParseUint(string(b), 10, 0)
    if err != nil {
        return err
    }

	if val == 0 {
		return errors.New("job count must be at least 1")
	}

	*jc = JobCount(val)
	return err
}

type appArgs struct {
	Inputs     string   `arg:"-i" default:"inputs.txt" help:"file with tests"`
	NoColors   bool     `arg:"--no-colors" help:"disable colored output"`
	NoProgress bool     `arg:"--no-progress" help:"disable progress bar"`
    Jobs       JobCount `arg:"-j" default:"CPU_COUNT" placeholder:"COUNT" help:"Number of tests to run concurrently"`
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
	return "cptest 2.02a"
}

func mustParse(dest *appArgs) {
	parser, err := arg.NewParser(arg.Config{}, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: couldn't initialize command line argument parser")
		os.Exit(-1)
	}

	cliArgs := os.Args[1:]

	for end := 0; end != len(cliArgs)+1; end++ {
		// Skip flags until we find a bare string, possibly the executable.
		// But let parser.Parse to execute at least once.
		if end != 0 && cliArgs[end-1][0] == '-' {
			continue
		}

		err = parser.Parse(cliArgs[:end])
		if err != nil {
			continue
		}

		dest.Args = cliArgs[end:]
		break
	}

	// It will handle help and version arguments.
	if err != nil {
		arg.MustParse(dest)
	}
}

func init() {
	mustParse(&args)

	if args.NoColors {
		cptest.Au = aurora.NewAurora(false)
	}

	verdictStr = map[cptest.Verdict]aurora.Value{
		cptest.OK: cptest.Au.Bold("OK").Green(),
		cptest.IE: cptest.Au.Bold("IE").Bold(),
		cptest.WA: cptest.Au.Bold("WA").Red(),
		cptest.RE: cptest.Au.Bold("RE").Magenta(),
		cptest.TL: cptest.Au.Bold("TL").Yellow(),
	}

	type duration cptest.PositiveDuration
	cptest.DefaultInputsConfig = cptest.InputsConfig{
		Tl:   cptest.NewPositiveDuration(6 * time.Second),
		Prec: 6,
	}
}

func main() {
	inputsPath, err := filepath.Abs(args.Inputs)
	if err != nil {
		fmt.Printf("error: retreive inputs absolute path: %v\n", err)
		return
	}

	inputs, scanErrs := readInputs(inputsPath)
	if scanErrs != nil {
		var lineRangeErrorType *cptest.LineRangeError
		if len(scanErrs) == 1 && !errors.As(scanErrs[0], &lineRangeErrorType) {
			fmt.Printf("error: %v\n", scanErrs[0])
			return
		}

		lineErrs := make([]*cptest.LineRangeError, len(scanErrs))
		for i, scanErr := range scanErrs {
			ok := errors.As(scanErr, &lineErrs[i])
			if !ok {
				panic(fmt.Sprintf("internal bug: some parse errors don't have line information"))
			}
		}

		sort.Slice(lineErrs, func(i, j int) bool {
			return lineErrs[i].Begin < lineErrs[j].Begin
		})

		errorHeader := cptest.Au.Bold("error").BrightRed()

		for _, err := range lineErrs {
			fmt.Printf("%s:%d: %s: %v\n%s", args.Inputs, err.Begin, errorHeader, err.Err, err.CodeSnippet())
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

	swatch := &cptest.ConfigurableStopwatcher{
		TL:    inputs.Config.Tl.Duration,
		Clock: clockwork.NewRealClock(),
	}
	pool := cptest.NewThreadPool(int(args.Jobs))

	batch := cptest.NewTestingBatch(inputs, proc, swatch, pool)

	if inputs.Config.Tl.Duration == 0 {
		fmt.Println("time limit: infinity")
	} else {
		fmt.Printf("time limit: %v\n", inputs.Config.Tl)
	}
	fmt.Printf("floating point precision: %d digit(s)\n", batch.Lx.Precision)
	fmt.Printf("job count: %d\n", args.Jobs)

	batch.TestEndCallback = verboseResultPrinter

	testingHeader := cptest.Au.Bold("    Testing").Cyan().String()

	progressBar = &ProgressBar{
		Total:  len(inputs.Tests),
		Width:  20,
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
