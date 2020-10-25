package cptest

import (
	"bytes"
	"io"
	"strings"
	"sync"
)

type Processer interface {
    GetError(id int) error
    GetOutput(id int) string

    Run(id int, in string) error
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

func (p *ProcessFunc) Run(id int, in string) error {
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

    return nil
}

func (p *ProcessFunc) WaitCompleted() int {
    return <-p.complete
}
