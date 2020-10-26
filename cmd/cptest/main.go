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

func ReadInputs(inputsPath string) (cptest.Inputs, []error) {
    inputsFile, err := os.Open(inputsPath)
    if err != nil {
        return cptest.Inputs{}, []error{fmt.Errorf("load tests: %v", err)}
    }
    defer inputsFile.Close()

    inputs, errs := cptest.ScanInputs(inputsFile)
    if errs != nil {
        for _, err := range errs {
            fmt.Printf("load tests: %v\n", err)
        }
        return cptest.Inputs{}, errs
    }

    return inputs, nil
}

func IsExec(filename string) error {
    info, err := os.Stat(filename)
    if err != nil {
        return fmt.Errorf("is executable: %v", err)
    }

    if info.IsDir() {
        return fmt.Errorf("is executable: %s is a directory", filename)
    }

    if info.Size() == 0 {
        return fmt.Errorf("is executable: %s is an empty file", filename)
    }

    if info.Mode()&0111 == 0 {
        return fmt.Errorf("%s is not an executable", filename)
    }

    return nil
}

func main() {

    flag.Parse()

    if len(flag.Args()) != 0 {
        wd = flag.Args()[0]
    }

    cwd, err := os.Getwd()
    if err != nil {
        fmt.Println("error: could not get path for the current working directory")
        return
    }

    wd = joinIfRelative(cwd, wd)
    inputsPath = joinIfRelative(wd, inputsPath)

    inputs, errs := ReadInputs(inputsPath)
    if errs != nil {
        for _, err := range errs {
            fmt.Println(err)
        }

        return
    }

    if execPath == "" {
        dir, err := os.Open(wd)
        if err != nil {
            fmt.Printf("error: search executable: %v\n", err)
            return
        }

        names, err := dir.Readdirnames(0)
        if err != nil {
            fmt.Printf("error: search executable: %v\n", err)
            return
        }

        var execs []string
        for _, name := range names {
            name = path.Join(wd, name)
            if IsExec(name) == nil {
                execs = append(execs, name)
            }
        }

        if len(execs) == 0 {
            fmt.Printf("error: no executables found in %s", wd)
            return
        }

        if len(execs) > 1 {
            fmt.Printf("error: more that one executable found in %s.", wd)
            fmt.Printf(" Choose appropriate one with -e flag.\nfound %d:\n", len(execs))
            
            for _, name := range execs {
                fmt.Println(name)
            }

            return
        }

        execPath = execs[0]
    }

    execPath = joinIfRelative(wd, execPath)
    err = IsExec(execPath)
    if err != nil {
        fmt.Printf("error: %v\n", err)
        return
    }

    proc := cptest.NewProcess(execPath)

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
