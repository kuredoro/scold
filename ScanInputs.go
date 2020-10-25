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

func ScanTest(str string) (Test, []error) {
    trueDelim := "\n" + IODelim + "\n"
    parts := strings.Split(str, trueDelim)

    if len(parts) < 2 {
        return Test{}, []error{fmt.Errorf("%w", NoSections)}
    }

    test := Test{
        Input: strings.TrimSpace(parts[0]),
        Output: strings.TrimSpace(strings.Join(parts[1:], trueDelim)),
    }

    return test, nil
}

func ScanInputs(r io.Reader) (Inputs, []error) {
    buf := &bytes.Buffer{}
    io.Copy(buf, r)

    test, errs := ScanTest(buf.String())

    if errs != nil {
        for i, err := range errs {
            errs[i] = fmt.Errorf("test 1: %w", err)
        }

        return Inputs{}, errs
    }

    var inputs Inputs
    inputs.Tests = []Test{test}

    return inputs, nil
}
