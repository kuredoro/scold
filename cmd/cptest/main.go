package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kuredoro/cptest"
)

var wd = "."

var inputsPath string
var execPath string

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			`CPTEST
        Competitive programming write/test cycle automation tool.

USAGE
        cptest [-i INPUTS] [-e EXECUTABLE] [WORKING_DIR]

FLAGS
`)
		flag.PrintDefaults()

		fmt.Fprintf(flag.CommandLine.Output(),
			`
EXECUTABLE SEARCHING
        If no executable path provided, cptest will try to find an executable
        inside working directory. If only one executable is found it is chosen
        as the executable to be tested.

INPUTS SYNTAX
        The input and output should be separated by 3 hyphes (---) on their own
        line. The input and output is stripped of spaces from both sides 
        line-wise. The input is given to the executable. The executable's
        output is also stripped of spaces from both sides and then compared with
        the expected output symbol by symbol. A valid test case should always
        contain the separator.

        Each individual test case is separated by 3 equality signs (===) on
        their own line. A test case may be empty (=== and another === on the
        next line). These are ignored.

        The first test case can describe a set of key-value pairs in format
        key=value. In this case, the key-value pairs will be used to fine-tune
        and configure cptest session. For example, a time limit can be changed
        to a custom one by providing "tl = 10.0" to set it to 10 seconds.

INPUTS EXAMPLE

        tl = 1
        ===
        5
        5 4 3 2 1
        ---
        1 2 3 4 5
        ===
        7
        2 1 4 3 6 5 7
        ---
        1 2 3 4 5 6 7

AUTHOR
        @kuredoro
        Project's GitHub page: https://github.com/kuredoro/cptest

        If you found a bug or would like to have some awesome feature, be sure
        to make an issue on GitHub!

VERSION
        1.0.0

`)
	}

	flag.StringVar(&inputsPath, "i", "inputs.txt", "File with the test cases (and, optionally, config)")
	flag.StringVar(&execPath, "e", "", "Path to the executable")
}

func GetProc() (proc cptest.Processer, err error) {
	if execPath == "" {
		execPath, err = FindExecutable(wd)

		if err != nil {
			return nil, err
		}

		fmt.Printf("found executable: %s\n", joinIfRelative(wd, execPath))
	}

	execPath = joinIfRelative(wd, execPath)

	/*
	   This check does not work on Windows.
	   TODO: Find fix.
	   if err = IsExec(execPath); err != nil {
	       return nil, err
	   }
	*/

	proc = &Executable{
		Path: execPath,
	}

	return
}

func main() {
	flag.Parse()

	if count := len(flag.Args()); count != 0 {
		wd = flag.Args()[0]

		if count > 1 {
			fmt.Printf("warning: expected 0 or 1 command line argument, got %v\n", count)
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("error: could not get path for current working directory")
		return
	}

	wd = joinIfRelative(cwd, wd)
	if err = CheckWd(wd); err != nil {
		fmt.Printf("error: %v\n", err)
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

	proc, err := GetProc()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	TL := GetTL(inputs)
	swatch := cptest.NewConfigurableStopwatcher(TL)

	batch := cptest.NewTestingBatch(inputs, proc, swatch)

	batch.ResultPrinter = VerboseResultPrinter

	batch.Run()

	passCount := 0
	for _, v := range batch.Stat {
		if v == cptest.OK {
			passCount++
		}
	}

	if passCount == len(batch.Stat) {
		fmt.Println("OK")
	} else {
		fmt.Printf("FAIL\n%d/%d passed\n", passCount, len(batch.Stat))
	}
}
