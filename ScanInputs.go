package cptest

import "io"

type Test struct {
    Input string
    Output string
}

type Inputs struct {
    Tests []Test
}

func ScanInputs(r io.Reader) Inputs {
    return Inputs{}
}
