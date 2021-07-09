package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kuredoro/cptest"
)

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
