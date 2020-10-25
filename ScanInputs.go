package cptest

import (
	"fmt"
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
