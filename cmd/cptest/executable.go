package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

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
        cleanStderr := strings.TrimSpace(string(ee.Stderr))
        fmt.Fprint(w, cleanStderr)
        return fmt.Errorf("%v", ee)
    }

    if err != nil {
        return fmt.Errorf("%w: %v", cptest.InternalErr, err)
    }

    var str strings.Builder
    s := bufio.NewScanner(bytes.NewReader(out))

    for s.Scan() {
        str.WriteString(strings.TrimSpace(s.Text()))
        str.WriteRune('\n')
    }

    fmt.Fprint(w, str.String())
    return nil
}
