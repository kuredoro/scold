package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/kuredoro/cptest"
)

type Executable struct {
	Path string
}

func trimSpaceLineWise(text string) string {
	var str strings.Builder
	s := bufio.NewScanner(strings.NewReader(text))

	for s.Scan() {
        line := strings.TrimSpace(s.Text())
        if line != "" {
            str.WriteString(strings.TrimSpace(s.Text()))
            str.WriteRune('\n')
        }
	}

	return str.String()
}

func (e *Executable) Run(r io.Reader, w io.Writer) error {
	cmd := exec.Command(e.Path)
	cmd.Stdin = r

	out, err := cmd.Output()

	if ee, ok := err.(*exec.ExitError); ok {
		cleanStderr := trimSpaceLineWise(string(ee.Stderr))
		fmt.Fprint(w, cleanStderr)
		return fmt.Errorf("%v", ee)
	}

	if err != nil {
		return fmt.Errorf("%w: %v", cptest.InternalErr, err)
	}

	cleanOut := trimSpaceLineWise(string(out))
	fmt.Fprint(w, cleanOut)
	return nil
}
