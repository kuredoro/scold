package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kuredoro/cptest"
	"github.com/logrusorgru/aurora"
)

var wd = "."

var inputsPath string
var execPath string

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
`CPTEST
        Feed apps fixed inputs, compare expected and their outputs.

USAGE
        cptest [-i INPUTS] EXECUTABLE

FLAGS
`)
		flag.PrintDefaults()

		fmt.Fprintf(flag.CommandLine.Output(),
			`
AUTHOR
        @kuredoro
        Usage guide: https://github.com/kuredoro/cptest

        Feature request or a bug report is always welcome at
        https://github.com/kuredoro/cptest/issues

VERSION
        1.1.0

`)
	}

	flag.StringVar(&inputsPath, "i", "inputs.txt", "File with the test cases")
}

func GetProc() (proc cptest.Processer, err error) {

	/*
	   This check does not work on Windows.
	   TODO: Find fix.
	   if err = IsExec(execPath); err != nil {
	       return nil, err
	   }
	*/

	return
}

func main() {
	flag.Parse()

	if count := len(flag.Args()); count != 1 {
		flag.Usage()

		return
	}

	execPath = flag.Args()[0]

	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("error: could not get path for current working directory")
		return
	}

	inputsPath = joinIfRelative(wd, inputsPath)

	inputs, errs := ReadInputs(inputsPath)
	if errs != nil {
		for _, err := range errs {
			fmt.Printf("error: %v\n", err)
		}

		return
	}

	execPath = joinIfRelative(wd, execPath)
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
		fmt.Println(aurora.Bold("OK").Green())
	} else {
		fmt.Println(aurora.Bold("FAIL").Red())
		fmt.Printf("%d/%d passed\n", passCount, len(batch.Verdicts))
	}
}
