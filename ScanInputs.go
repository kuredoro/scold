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
    NoSections = InputsError("no io separator found")
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

    trueDelim := "\n" + IODelim + "\n"
    parts := strings.Split(buf.String(), trueDelim)

    if len(parts) < 2 {
        return Inputs{}, []error{fmt.Errorf("%w", NoSections)}
    }

    test := Test{
        Input: strings.TrimSpace(parts[0]),
        Output: strings.TrimSpace(strings.Join(parts[1:], trueDelim)),
    }

    return Inputs{
        Tests: []Test{test},
    }, nil
}
