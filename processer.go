package cptest

import (
	"bytes"
	"io"
	"strings"
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

    errs map[int]error
    outs map[int]string
}

func NewProcessFunc(f IOFunc) *ProcessFunc {
    return &ProcessFunc{
        f: f,
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
    buf := &bytes.Buffer{}
    p.f(strings.NewReader(in), buf)
    
    p.errs[id] = nil
    p.outs[id] = strings.TrimSpace(buf.String())

    return nil
}

func (p *ProcessFunc) WaitCompleted() int {
    for id := range p.errs {
        return id
    }

    return 0
}
