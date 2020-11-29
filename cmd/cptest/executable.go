package main

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/kuredoro/cptest"
)

type Executable struct {
	Path string
}

func (e *Executable) Run(r io.Reader, w io.Writer) error {
	cmd := exec.Command(e.Path)
	cmd.Stdin = r

	out, err := cmd.Output()

	if ee, ok := err.(*exec.ExitError); ok {
		fmt.Fprint(w, string(ee.Stderr))
		return fmt.Errorf("%v", ee)
	}

	if err != nil {
		return fmt.Errorf("%w: %v", cptest.InternalErr, err)
	}

	fmt.Fprint(w, string(out))
	return nil
}
