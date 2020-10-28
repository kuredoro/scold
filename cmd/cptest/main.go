package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/kureduro/cptest"
)

const defaultTL = 6 * time.Second

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
        for i, err := range errs {
            errs[i] = fmt.Errorf("load tests: %v", err)
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

func FindExecutable(dirPath string) (string, error) {
    dir, err := os.Open(wd)
    if err != nil {
        return "", fmt.Errorf("search executable: %v\n", err)
    }

    names, err := dir.Readdirnames(0)
    if err != nil {
        return "", fmt.Errorf("search executable: %v\n", err)
    }

    var execs []string
    for _, name := range names {
        name = path.Join(wd, name)
        if IsExec(name) == nil {
            execs = append(execs, name)
        }
    }

    if len(execs) == 0 {
        return "", fmt.Errorf("no executables found in %s", wd)
    }

    if len(execs) > 1 {
        var msg strings.Builder

        msg.WriteString(fmt.Sprintf("error: more that one executable found in %s. ", wd))
        msg.WriteString("Choose appropriate one with -e flag.\n")
        msg.WriteString(fmt.Sprintf("found %d:\n", len(execs)))
        
        for _, name := range execs {
            msg.WriteString(name)
            msg.WriteRune('\n')
        }

        return "", errors.New(msg.String())
    }

    return execs[0], nil
}

func CheckWd(wd string) error {
    _, err := os.Open(wd)
    if err != nil {
        return fmt.Errorf("check working directory: %v", err)
    }

    return nil
}

func GetTL(inputs cptest.Inputs) (TL time.Duration) {
    TL = defaultTL

    if str, exists := inputs.Config["tl"]; exists {
        sec, err := strconv.ParseFloat(str, 64)
        if err != nil {
            fmt.Printf("warning: time limit %q is of incorrect format: %v\n", str, err)
            return
        }

        TL = time.Duration(sec * float64(time.Second))
        return
    }

    fmt.Printf("using default time limit: %v\n", defaultTL)
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
        fmt.Println("error: could not get path for the current working directory")
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

    if execPath == "" {
        execPath, err = FindExecutable(wd)

        if err != nil {
            fmt.Printf("error: %v", err)
            return
        }

        fmt.Printf("found executable: %s\n", joinIfRelative(wd, execPath))
    }

    execPath = joinIfRelative(wd, execPath)

    if err = IsExec(execPath); err != nil {
        fmt.Printf("error: %v\n", err)
        return
    }

    proc := &cptest.Executable{
        Path: execPath,
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
        fmt.Printf("FAIL\t(%d/%d passed)\n", passCount, len(batch.Stat))
    }
}
