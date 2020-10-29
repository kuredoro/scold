package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kureduro/cptest"
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
NOTES
        If no executable path provided, cptest will try to find an executable
        inside working directory. If only one executable is found it is chosen
        as the executable to be tested.
`)
    }

    flag.StringVar(&inputsPath, "i", "inputs.txt", "File with the test cases (and, optionally, config)")
    flag.StringVar(&execPath, "e", "", "Path to the executable")
}

func GetProc() (proc cptest.Processer, err error) {
    var path string

    if execPath == "" {
        path, err = FindExecutable(wd)

        if err != nil {
            return nil, err
        }

        fmt.Printf("found executable: %s\n", joinIfRelative(wd, path))
    }

    path = joinIfRelative(wd, path)

    if err = IsExec(path); err != nil {
        return nil, err
    }

    proc = &cptest.Executable{
        Path: path,
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
