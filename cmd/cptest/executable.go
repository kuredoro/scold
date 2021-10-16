package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/kuredoro/cptest"
)

type Executable struct {
	Path string
	Args []string
}

func (e *Executable) Run(ctx context.Context, r io.Reader) (cptest.ExecutionResult, error) {
	cmd := exec.Command(e.Path, e.Args...)
	cmd.Stdin = r

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return cptest.ExecutionResult{}, fmt.Errorf("executable: %v", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return cptest.ExecutionResult{}, fmt.Errorf("executable: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		return cptest.ExecutionResult{}, fmt.Errorf("executable: %v", err)
	}

	stdout := make([]byte, 0, 1024)
	stderr := make([]byte, 0, 1024)
	stdoutComplete := make(chan error)
	stderrComplete := make(chan error)

	go listenPipe(stdoutPipe, &stdout, stdoutComplete)
	go listenPipe(stderrPipe, &stderr, stderrComplete)

	for doneCount := 0; doneCount != 2; {
		select {
		case <-ctx.Done():
			// When process is killed the pipes are closed. the listenPipes
			// will receive EOF and return nil.
            err = cmd.Process.Kill()
            if err != nil {
                return cptest.ExecutionResult{}, fmt.Errorf("executable: kill: %v", err)
            }
		case err := <-stdoutComplete:
			if err != nil {
				return cptest.ExecutionResult{}, fmt.Errorf("executable: stdout: %v", err)
			}
			doneCount++
		case err := <-stderrComplete:
			if err != nil {
				return cptest.ExecutionResult{}, fmt.Errorf("executable: stderr: %v", err)
			}
			doneCount++
		}
	}

	close(stdoutComplete)
	close(stderrComplete)

	out := cptest.ExecutionResult{
		ExitCode: 0,
		Stdout:   string(stdout),
		Stderr:   string(stderr),
	}

	err = cmd.Wait()

	if ee, ok := err.(*exec.ExitError); ok {
		out.ExitCode = ee.ExitCode()
		return out, nil
	}

	if err != nil {
		return cptest.ExecutionResult{}, fmt.Errorf("executable: %v", err)
	}

	return out, nil
}

func listenPipe(pipe io.Reader, out *[]byte, done chan error) {
	buf := make([]byte, 1024)
	for {
		n, err := pipe.Read(buf)
		*out = append(*out, buf[:n]...)

		if err != nil && err != io.EOF {
			done <- err
			return
		}

		if err == io.EOF {
			done <- nil
			return
		}
	}
}
