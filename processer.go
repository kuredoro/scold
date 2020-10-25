package cptest

import (
    "io"
)

type Processer interface {
    SetInput(r io.Reader)
    GetOutput() []byte
    Run() error
}

type IOFunc func(io.Reader, io.Writer)

type ProcessFunc struct {

}

func NewProcessFunc(f IOFunc) *ProcessFunc {
    return nil
}

func (p *ProcessFunc) SetInput(r io.Reader) {

}

func (p *ProcessFunc) GetOutput() []byte {
    return nil
}

func (p *ProcessFunc) Run() error {
    return nil
}
