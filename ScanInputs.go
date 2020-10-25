package cptest

import (
	"bytes"
	"io"
	"strings"
)

const IODelim = "---"

type Test struct {
    Input string
    Output string
}

type Inputs struct {
    Tests []Test
}

func ScanInputs(r io.Reader) Inputs {
    buf := &bytes.Buffer{}
    io.Copy(buf, r)

    parts := strings.Split(buf.String(), "\n" + IODelim + "\n")

    test := Test{
        Input: strings.TrimSpace(parts[0]),
        Output: strings.TrimSpace(parts[1]),
    }

    return Inputs{
        Tests: []Test{test},
    }
}
