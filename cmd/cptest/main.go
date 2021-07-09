package main

import (
	"fmt"
    "path/filepath"

	"github.com/alexflint/go-arg"
	"github.com/kuredoro/cptest"
	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-colorable"
)

var stdout = colorable.NewColorableStdout()

type appArgs struct {
	Inputs     string `arg:"-i" default:"inputs.txt" help:"file with tests"`
	NoColors   bool   `arg:"--no-colors" help:"disable colored output"`
	Executable string `arg:"positional,required"`
    Args   []string `arg:"positional"`
}

var args appArgs

func (appArgs) Description() string {
	return `Feed programs fixed inputs, compare their outputs against expected ones.

Author: @kuredoro
User manual: https://github.com/kuredoro/cptest
`
}

func (appArgs) Version() string {
	return "cptest 2.01a"
}

func init() {
	arg.MustParse(&args)

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
}

func main() {
	inputsPath, err := filepath.Abs(args.Inputs)
    if err != nil {
        fmt.Printf("error: retreive inputs absolute path: %v", err)
        return
    }

	inputs, errs := readInputs(inputsPath)
	if errs != nil {
		for _, err := range errs {
			fmt.Printf("error: %v\n", err)
		}

		return
	}

	execPath, err := filepath.Abs(args.Executable)
    if err != nil {
        fmt.Printf("error: retreive executable absolute path: %v", err)
        return
    }

	proc := &Executable{
		Path: execPath,
        Args: args.Args,
	}

	TL := getTL(inputs)
	swatch := cptest.NewConfigurableStopwatcher(TL)

	batch := cptest.NewTestingBatch(inputs, proc, swatch)

	fmt.Printf("floating point precision: %d digit(s)\n", batch.Lx.Precision)

	batch.TestStartCallback = runPrinter
	batch.TestEndCallback = verboseResultPrinter

	batch.Run()

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
		fmt.Printf("%d/%d passed\n", passCount, len(batch.Verdicts))
	}
}
