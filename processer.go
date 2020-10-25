package cptest

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

type InternalError struct{}

func (e InternalError) Error() string {
    return "internal:"
}

var internalErr InternalError


type Processer interface {
    GetError(id int) error
    GetOutput(id int) string

    Run(id int, in string)
    WaitCompleted() int
}

type IOFunc func(io.Reader, io.Writer)

type ProcessFunc struct {
    f IOFunc

    complete chan int

    mu sync.Mutex
    errs map[int]error
    outs map[int]string

    Completed []int
}

func NewProcessFunc(f IOFunc) *ProcessFunc {
    return &ProcessFunc{
        f: f,
        complete: make(chan int),
        errs: make(map[int]error),
        outs: make(map[int]string),
    }
}

func (p *ProcessFunc) GetError(id int) error {
    return p.errs[id]
}

func (p *ProcessFunc) GetOutput(id int) string {
    return p.outs[id]
}

func (p *ProcessFunc) Run(id int, in string) {
    go func() {
        buf := &bytes.Buffer{}
        p.f(strings.NewReader(in), buf)

        p.mu.Lock()
        defer p.mu.Unlock()

        p.outs[id] = strings.TrimSpace(buf.String())
        p.errs[id] = nil
        p.Completed = append(p.Completed, id)

        p.complete<-id
    }()
}

func (p *ProcessFunc) WaitCompleted() int {
    return <-p.complete
}


type Process struct {
    path string

    complete chan int

    mu sync.Mutex
    errs map[int]error
    outs map[int]string
}

func (p *Process) NewProcess(execPath string) *Process {
    return &Process{
        path: execPath,
        complete: make(chan int),
        errs: make(map[int]error),
        outs: make(map[int]string),
    }
}

func (p *Process) GetError(id int) error {
    return p.errs[id]
}

func (p *Process) GetOutput(id int) string {
    return p.outs[id]
}

func (p *Process) Run(id int, in string) {
    go func() {
        defer func() { 
            p.complete<-id 
        }()

        cmd := exec.Command(p.path)
        cmd.Stdin = strings.NewReader(in)

        out, err := cmd.Output()

        p.mu.Lock()
        defer p.mu.Unlock()

        if err != nil {
            p.errs[id] = fmt.Errorf("%v", err)
            ee, ok := err.(*exec.ExitError)

            if !ok {
                p.errs[id] = fmt.Errorf("%w: %v", internalErr, err)
                p.outs[id] = ""
                return
            }

            p.outs[id] = string(ee.Stderr)
            return
        }

        p.errs[id] = nil
        p.outs[id] = strings.TrimSpace(string(out))
    }()
}

func (p *Process) WaitCompleted() int {
    return <-p.complete
}
