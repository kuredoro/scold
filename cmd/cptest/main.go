package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/kuredoro/cptest"
	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-colorable"
)

var wd = "."

var stdout = colorable.NewColorableStdout()

type appArgs struct {
	Inputs     string `arg:"-i" default:"inputs.txt" help:"file with tests"`
	Executable string `arg:"positional,required"`
	NoColors   bool   `arg:"--no-colors" help:"disable colored output"`
}

var args appArgs

func (appArgs) Description() string {
	return `Feed programs fixed inputs, compare their outputs against expected ones.

Author: @kuredoro
User manual: https://github.com/kuredoro/cptest
`
}

func (appArgs) Version() string {
	return "cptest 1.02a"
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
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("error: could not get path for current working directory")
		return
	}

	inputsPath := joinIfRelative(wd, args.Inputs)

	inputs, errs := readInputs(inputsPath)
	if errs != nil {
		for _, err := range errs {
			fmt.Printf("error: %v\n", err)
		}

		return
	}

	execPath := joinIfRelative(wd, args.Executable)
	proc := &Executable{
		Path: execPath,
	}
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
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
