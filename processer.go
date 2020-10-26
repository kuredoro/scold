package cptest

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Processer interface {
    Run(io.Reader, io.Writer) error
}


type SpyProcesser struct {
    Proc Processer

    CallCount int
}

func (p *SpyProcesser) Run(r io.Reader, w io.Writer) error {
    p.CallCount++
    return p.Proc.Run(r, w)
}


type ProcesserFunc func(io.Reader, io.Writer) error

func (p ProcesserFunc) Run(r io.Reader, w io.Writer) error {
    return p(r, w)
}


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
        return fmt.Errorf("%w: %v", internalErr, err)
    }

    cleanOutput := strings.TrimSpace(string(out))
    fmt.Fprint(w, cleanOutput)
    return nil
}
