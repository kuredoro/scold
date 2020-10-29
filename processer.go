package cptest

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Processer interface abstracts away the concept of the executable under
// testing.
type Processer interface {
    Run(io.Reader, io.Writer) error
}

// SpyProcesser is a test double that proxies another processer.
// It additionally stores the number of calls made to the Run function.
type SpyProcesser struct {
    Proc Processer

    CallCount int
}

// Run will execute the Run function of the inner processer, but will
// also increase the call count by one.
func (p *SpyProcesser) Run(r io.Reader, w io.Writer) error {
    p.CallCount++
    return p.Proc.Run(r, w)
}

// ProcesserFunc represents an implementation of Processer that instead
// of a real OS-level process executes Go code.
type ProcesserFunc func(io.Reader, io.Writer) error

// Run will call the underlying Go function to compute the result.
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
