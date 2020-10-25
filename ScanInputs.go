package cptest

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type InputsError string

func (e InputsError) Error() string {
    return string(e)
}

const (
    TooManySections = InputsError("too many sections")
)

const IODelim = "---"

type Test struct {
    Input string
    Output string
}

type Inputs struct {
    Tests []Test
}

func ScanInputs(r io.Reader) (Inputs, []error) {
    buf := &bytes.Buffer{}
    io.Copy(buf, r)

    parts := strings.Split(buf.String(), "\n" + IODelim + "\n")

    if len(parts) > 2 {
        return Inputs{}, []error{fmt.Errorf("%w", TooManySections)}
    }

    test := Test{
        Input: strings.TrimSpace(parts[0]),
        Output: strings.TrimSpace(parts[1]),
    }

    return Inputs{
        Tests: []Test{test},
    }, nil
}
