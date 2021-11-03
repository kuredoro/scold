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
	"github.com/kuredoro/scold"
	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-colorable"
    "github.com/mattn/go-isatty"
)

var progressBar *ProgressBar

var stdout = colorable.NewColorableStdout()

var errorLabel aurora.Value

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
	Inputs        string   `arg:"-i" default:"inputs.txt" help:"file with tests"`
	NoColors      bool     `arg:"--no-colors" help:"disable colored output"`
    ForceColors   bool     `arg:"--force-colors" help:"print colors even in non-tty contexts"`
	NoProgress    bool     `arg:"--no-progress" help:"disable progress bar"`
    ForceProgress bool     `arg:"--force-progress" help:"print progress bar even in non-tty contexts"`
	Jobs          JobCount `arg:"-j" default:"CPU_COUNT" placeholder:"COUNT" help:"Number of tests to run concurrently"`
	Executable    string   `arg:"positional,required"`
	Args          []string `arg:"positional" placeholder:"ARG"`
}

var args appArgs

func (appArgs) Description() string {
	return `Feed programs fixed inputs, compare their outputs against expected ones.

Author: @kuredoro (github, twitter)
User manual: https://github.com/kuredoro/scold
`
}

func (appArgs) Version() string {
	return "scold 2.03a"
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

	if args.NoColors && args.ForceColors {
        fmt.Printf("warning: colors are forced and disabled at the same time. --no-colors is always preferred.")
	}

	if args.NoProgress && args.ForceProgress {
        fmt.Printf("warning: progress bar is forced and disabled at the same time. --no-progress is always preferred.")
	}

    fd := os.Stdout.Fd()
    istty := isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)

    if !args.ForceColors && !istty {
        args.NoColors = true
    }

    if !args.ForceProgress && !istty {
        args.NoProgress = true
    }

	if args.NoColors {
		scold.Au = aurora.NewAurora(false)
	}

	verdictStr = map[scold.Verdict]aurora.Value{
		scold.OK: scold.Au.Bold("OK").Green(),
		scold.IE: scold.Au.Bold("IE").Bold(),
		scold.WA: scold.Au.Bold("WA").BrightRed(),
		scold.RE: scold.Au.Bold("RE").Magenta(),
		scold.TL: scold.Au.Bold("TL").Yellow(),
	}

	errorLabel = scold.Au.Bold("error").BrightRed()

	type duration scold.PositiveDuration
	scold.DefaultInputsConfig = scold.InputsConfig{
		Tl:   scold.NewPositiveDuration(6 * time.Second),
		Prec: 6,
	}
}

func main() {
	inputsPath, err := filepath.Abs(args.Inputs)
	if err != nil {
		fmt.Printf("%v: retreive inputs absolute path: %v\n", errorLabel, err)
		return
	}

	inputs, scanErrs := readInputs(inputsPath)
	if scanErrs != nil {
		var lineRangeErrorType *scold.LineRangeError
		if len(scanErrs) == 1 && !errors.As(scanErrs[0], &lineRangeErrorType) {
			fmt.Printf("%v: %v\n", errorLabel, scanErrs[0])
			return
		}

		lineErrs := make([]*scold.LineRangeError, len(scanErrs))
		for i, scanErr := range scanErrs {
			ok := errors.As(scanErr, &lineErrs[i])
			if !ok {
				panic(fmt.Sprintf("internal bug: some parse errors don't have line information"))
			}
		}

		sort.Slice(lineErrs, func(i, j int) bool {
			return lineErrs[i].Begin < lineErrs[j].Begin
		})

		for _, err := range lineErrs {
			fmt.Printf("%s:%d: %s: %v\n%s", args.Inputs, err.Begin, errorLabel, err.Err, err.CodeSnippet())
		}

		return
	}

	execPath, err := findFile(args.Executable)
	if err != nil {
		fmt.Printf("%v: find executable: %v\n", errorLabel, err)
		return
	}

	execStat, err := os.Stat(execPath)
	if err != nil {
		fmt.Printf("%v: read executable's properties: %v\n", errorLabel, err)
		return
	}

	if execStat.IsDir() {
		fmt.Printf("%v: provided executable %s is a directory\n", errorLabel, args.Executable)
		return
	}

	proc := &Executable{
		Path: execPath,
		Args: args.Args,
	}

	swatch := &scold.ConfigurableStopwatcher{
		TL:    inputs.Config.Tl.Duration,
		Clock: clockwork.NewRealClock(),
	}
	pool := scold.NewThreadPool(int(args.Jobs))

	batch := scold.NewTestingBatch(inputs, proc, swatch, pool)

	if inputs.Config.Tl.Duration == 0 {
		fmt.Println("time limit: infinity")
	} else {
		fmt.Printf("time limit: %v\n", inputs.Config.Tl)
	}
	fmt.Printf("floating point precision: %d digit(s)\n", batch.Lx.Precision)
	fmt.Printf("job count: %d\n", args.Jobs)

	batch.TestEndCallback = verboseResultPrinter

	testingHeader := scold.Au.Bold("    Testing").Cyan().String()

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
	for _, r := range batch.Results {
		if r.Verdict == scold.OK {
			passCount++
		}
	}

	if passCount == len(batch.Results) {
		fmt.Fprintln(stdout, scold.Au.Bold("OK").Green())
	} else {
		fmt.Fprintln(stdout, scold.Au.Bold("FAIL").Red())
		fmt.Fprintf(stdout, "%d/%d passed\n", passCount, len(batch.Results))
	}
}
