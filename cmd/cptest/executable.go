package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"

	"github.com/kuredoro/cptest"
)

type Executable struct {
	Path string
}

func (e *Executable) Run(r io.Reader) (cptest.ProcessResult, error) {
	cmd := exec.Command(e.Path)
	cmd.Stdin = r

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return cptest.ProcessResult{}, fmt.Errorf("executable: %v", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return cptest.ProcessResult{}, fmt.Errorf("executable: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		return cptest.ProcessResult{}, fmt.Errorf("executable: %v", err)
	}

	stdErr, err := ioutil.ReadAll(stderrPipe)
	if err != nil {
		return cptest.ProcessResult{}, fmt.Errorf("executable: %v", err)
	}

	stdOut, err := ioutil.ReadAll(stdoutPipe)
	if err != nil {
		return cptest.ProcessResult{}, fmt.Errorf("executable: %v", err)
	}

	out := cptest.ProcessResult{
		ExitCode: 0,
		Stdout:   string(stdOut),
		Stderr:   string(stdErr),
	}

	err = cmd.Wait()

	if ee, ok := err.(*exec.ExitError); ok {
		out.ExitCode = ee.ExitCode()
		return out, nil
	}

	if err != nil {
		return cptest.ProcessResult{}, fmt.Errorf("executable: %v", err)
	}

	return out, nil
}
