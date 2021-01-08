package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/kuredoro/cptest"
	"github.com/logrusorgru/aurora"
)

var wd = "."

type args struct {
    Inputs string `arg:"-i" default:"inputs.txt" help:"file with tests"`
    Executable string `arg:"positional,required"`
    NoColors bool `arg:"--no-colors" help:"disable colored output"`
}

var Args args

func (args) Description() string {
    return `Feed programs fixed inputs, compare their outputs against expected ones.

Author: @kuredoro
User manual: https://github.com/kuredoro/cptest
`
}

func (args) Version() string {
    return "cptest 1.01z"
}

func init() {
    arg.MustParse(&Args)

    if Args.NoColors {
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

    inputsPath := joinIfRelative(wd, Args.Inputs)

	inputs, errs := ReadInputs(inputsPath)
	if errs != nil {
		for _, err := range errs {
			fmt.Printf("error: %v\n", err)
		}

		return
	}

    execPath := joinIfRelative(wd, Args.Executable)
	proc := &Executable{
		Path: execPath,
	}
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	TL := GetTL(inputs)
	swatch := cptest.NewConfigurableStopwatcher(TL)

	batch := cptest.NewTestingBatch(inputs, proc, swatch)

    fmt.Printf("floating point precision: %d digit(s)\n", batch.Lx.Precision)

	batch.TestStartCallback = RunPrinter
	batch.TestEndCallback = VerboseResultPrinter

	batch.Run()

	passCount := 0
	for _, v := range batch.Verdicts {
		if v == cptest.OK {
			passCount++
		}
	}

	if passCount == len(batch.Verdicts) {
		fmt.Println(cptest.Au.Bold("OK").Green())
	} else {
		fmt.Println(cptest.Au.Bold("FAIL").Red())
		fmt.Printf("%d/%d passed\n", passCount, len(batch.Verdicts))
	}
}
