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
	"time"

	"github.com/alexflint/go-arg"
	"github.com/jonboulle/clockwork"
	"github.com/kuredoro/scold"
	"github.com/kuredoro/scold/forwarders"
	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

var stdout = colorable.NewColorableStdout()

var errorLabel, warningLabel aurora.Value

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

func errorPrintf(format string, args ...any) {
    printArgs := make([]any, len(args) + 1)
    printArgs[0] = errorLabel
    copy(printArgs[1:], args)

    fmt.Fprintf(stdout, "%v: " + format + "\n", printArgs...)
}

func warningPrintf(format string, args ...any) {
    printArgs := make([]any, len(args) + 1)
    printArgs[0] = warningLabel
    copy(printArgs[1:], args)

    fmt.Fprintf(stdout, "%v: " + format + "\n", printArgs...)
}

func init() {
	mustParse(&args)

	if args.NoColors && args.ForceColors {
        fmt.Println("warning: colors are forced and disabled at the same time. --no-colors is always preferred.")
	}

	if args.NoProgress && args.ForceProgress {
        fmt.Println("warning: progress bar is forced and disabled at the same time. --no-progress is always preferred.")
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

	errorLabel = scold.Au.Bold("error").BrightRed()
	warningLabel = scold.Au.Bold("warning").BrightYellow()

	scold.DefaultInputsConfig = scold.InputsConfig{
		Tl:   scold.NewPositiveDuration(6 * time.Second),
		Prec: 8,
	}
}

func main() {
	inputsPath, err := filepath.Abs(args.Inputs)
	if err != nil {
		errorPrintf("retreive inputs absolute path: %v", err)
        os.Exit(1)
	}

	inputs, scanErrs := readInputs(inputsPath)
	if scanErrs != nil {
		var lineRangeErrorType *scold.LineRangeError
		if len(scanErrs) == 1 && !errors.As(scanErrs[0], &lineRangeErrorType) {
            errorPrintf("load tests: %v", scanErrs[0])
            os.Exit(1)
		}

		lineErrs := make([]*scold.LineRangeError, len(scanErrs))
		for i, scanErr := range scanErrs {
			ok := errors.As(scanErr, &lineErrs[i])
			if !ok {
                panic(fmt.Sprintf("internal bug: some parse errors don't have line information (%v)", scanErr))
			}
		}

		sort.Slice(lineErrs, func(i, j int) bool {
			return lineErrs[i].Begin < lineErrs[j].Begin
		})

        hadErrors := false
		for _, err := range lineErrs {
            if w := scold.StringWarning(""); errors.As(err, &w) {
                warningPrintf("%s:%d: %v\n%s", args.Inputs, err.Begin, err.Err, err.CodeSnippet())
            } else {
                errorPrintf("%s:%d: %v\n%s", args.Inputs, err.Begin, err.Err, err.CodeSnippet())
                hadErrors = true
            }
		}

        if hadErrors {
            os.Exit(1)
        }
	}

	execPath, err := findFile(args.Executable)
	if err != nil {
		errorPrintf("find executable: %v", err)
        os.Exit(1)
	}

	execStat, err := os.Stat(execPath)
	if err != nil {
		errorPrintf("read executable's properties: %v", err)
        os.Exit(1)
	}

	if execStat.IsDir() {
		errorPrintf("provided executable %s is a directory", args.Executable)
        os.Exit(1)
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

	var progressBar *ProgressBar
	if !args.NoProgress {
		testingHeader := scold.Au.Bold("    Testing").Cyan().String()
		progressBar = &ProgressBar{
			Total:  len(inputs.Tests),
			Width:  20,
			Header: testingHeader,
		}
	}

	cliPrinter := NewPrettyPrinter(scold.Au)
	cliPrinter.Bar = progressBar

	asyncF := forwarders.NewAsyncEventForwarder(cliPrinter, 100)
	batch.Listener = asyncF

	batch.Run()

	asyncF.Wait()

    allOK := true
    for i := range batch.Results {
        if batch.Results[i].Verdict != scold.OK {
            allOK = false
            break
        }
    }

    if !allOK {
        os.Exit(1)
    }
}
