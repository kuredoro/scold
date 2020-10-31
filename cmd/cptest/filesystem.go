package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kuredoro/cptest"
)

func joinIfRelative(dir, file string) string {
    if file != "" && file[0] == '/' {
        return file
    }

    return filepath.Join(dir, file)
}

func ReadInputs(inputsPath string) (cptest.Inputs, []error) {
    inputsFile, err := os.Open(inputsPath)
    if err != nil {
        return cptest.Inputs{}, []error{fmt.Errorf("load tests: %v", err)}
    }
    defer inputsFile.Close()

    text, err := ioutil.ReadAll(inputsFile)
    if err != nil {
        return cptest.Inputs{}, []error{fmt.Errorf("load tests: %v", err)}
    }

    inputs, errs := cptest.ScanInputs(string(text))
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
        name = filepath.Join(wd, name)
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
