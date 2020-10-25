package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/kureduro/cptest"
)

var wd = "."

var inputsPath string
var execPath string

func joinIfRelative(dir, filepath string) string {
    if filepath != "" && filepath[0] == '/' {
        return filepath
    }

    return path.Join(dir, filepath)
}

func init() {
    flag.StringVar(&inputsPath, "i", "inputs.txt", "File with test cases ")
    flag.StringVar(&execPath, "e", "", "Path to the executable")
}

func main() {

    flag.Parse()

    if len(flag.Args()) != 0 {
        wd = flag.Args()[0]
    }

    inputsPath = joinIfRelative(wd, inputsPath)
    execPath = joinIfRelative(wd, execPath)

    inputsFile, err := os.Open(inputsPath)
    if err != nil {
        fmt.Printf("error: %v\n", err)
        return
    }
    defer inputsFile.Close()

    inputs, errs := cptest.ScanInputs(inputsFile)
    if errs != nil {
        for _, err := range errs {
            fmt.Printf("error: %v\n", err)
        }
        return
    }

    proc := cptest.NewProcess(execPath)
    if err != nil {
        fmt.Printf("error: %v\n", err)
        return
    }

    batch := cptest.NewTestingBatch(inputs, proc)

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
        fmt.Printf("FAIL\t(%d/%d passed)\n", passCount, len(batch.Stat))
    }
}
