package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kuredoro/cptest"
)

func getPath() []string {
	delim := ":"
	if runtime.GOOS == "windows" {
		delim = ";"
	}

	list := os.Getenv("PATH")
	return strings.Split(list, delim)
}

func findFile(userPath string) (string, error) {
	// Ability to omit .exe prefix on Windows.
	if runtime.GOOS == "windows" && filepath.Ext(userPath) == "" {
		userPath += ".exe"
	}

	absPath, err := filepath.Abs(userPath)
	pathForError := absPath
	if err != nil {
		fmt.Printf("warning: could not retrieve executables absolute path: %v. Will look in PATH", err)
		// For error report at the end of the func
		pathForError = userPath
	}

	var candidates []string
	if err == nil {
		candidates = append(candidates, absPath)
	}

	base := filepath.Base(userPath)

	// If only the file name supplied, search in PATH
	if filepath.Dir(userPath) == "." {
		extraPaths := getPath()
		for _, path := range extraPaths {
			candidates = append(candidates, filepath.Join(path, base))
		}
	}

	for _, cand := range candidates {
		if _, err := os.Stat(cand); err == nil {
			return cand, nil
		}
	}

	if filepath.Dir(userPath) == "." {
		return "", fmt.Errorf("%s is absent in current working directory and in PATH", userPath)
	}

	return "", fmt.Errorf("%s does not exist", pathForError)
}

func readInputs(inputsPath string) (cptest.Inputs, []error) {
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
