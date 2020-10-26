package cptest

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type InputsError string

func (e InputsError) Error() string {
    return string(e)
}

const (
    IOSeparatorMissing = InputsError("IO separator missing")
    KeyMissing = InputsError("key cannot be empty")
    ValueMissing = InputsError("value cannot be empty")
    KVMissing = InputsError("key and value are missing")
    NotKVPair = InputsError("not a key-value pair")
)

type LinedError struct {
    Header string
    Line int
    Err error
}

func (e *LinedError) Error() string {
    return fmt.Sprintf("%s: line %d: %v", e.Header, e.Line, e.Err)
}

func (e *LinedError) Unwrap() error {
    return e.Err
}


const (
    IODelim = "---"
    TestDelim = "==="
)

type Test struct {
    Input string
    Output string
}

type Inputs struct {
    Tests []Test
    Config map[string]string
}

func ScanTest(str string) (Test, []error) {
    if strings.TrimSpace(str) == "" || str == IODelim {
        return Test{}, nil
    }

    trueDelim := "\n" + IODelim + "\n"
    parts := strings.SplitN(str, trueDelim, 2)

    if len(parts) < 2 {
        // Maybe --- is on first line
        trueDelim = IODelim + "\n"
        parts = strings.SplitN(str, trueDelim, 2)

        if len(parts) < 2 || parts[0] != "" {
            return Test{}, []error{fmt.Errorf("%w", IOSeparatorMissing)}
        }

    }

    test := Test{
        Input: strings.TrimSpace(parts[0]),
        Output: strings.TrimSpace(parts[1]),
    }

    return test, nil
}

func splitByString(data []byte, atEOF bool, delim string) (advance int, token []byte, err error) {

    trueDelim := delim + "\n"

    if len(trueDelim) <= len(data) {
        prefix := data[:len(trueDelim)]

        if string(prefix) == trueDelim {
            return len(trueDelim), []byte{}, nil
        }
    }

    trueDelim = "\n" + delim + "\n"

    prefixEnd := 0
    for i := 0; i < len(data); i++ {
        if data[i] == trueDelim[prefixEnd] {
            prefixEnd++
        } else {
            prefixEnd = 0
        }

        if prefixEnd == len(trueDelim) {
            return i + 1, data[:i-prefixEnd+1], nil
        }
    }

    if atEOF && len(data) != 0 {
        testLen := len(data)

        // Explicit check that we have === at the very end with no \n at the end
        trueDelim = "\n" + delim
        if len(trueDelim) <= len(data) {
            suffix := data[len(data)-len(trueDelim):]

            if string(suffix) == trueDelim {
                testLen = len(data) - len(trueDelim)
            }
        }

        return testLen, data[:testLen], nil
    }

    return 0, nil, nil
}


func ScanKeyValuePair(line string) (string, string, error) {
    parts := strings.Split(line, "=")

    if len(parts) == 1 {
        cleanLine := strings.TrimSpace(line)

        if cleanLine == "" {
            return "", "", nil
        }

        if cleanLine == "=" {
            return "", "", KVMissing
        }

        return "", "", NotKVPair
    }

    if parts[0] == "" {
        return "", "", KeyMissing
    }

    if parts[1] == "" {
        return "", "", ValueMissing
    }

    key := strings.TrimSpace(parts[0])
    val := strings.TrimSpace(strings.Join(parts[1:], "="))

    return key, val, nil
}

func ScanConfig(text string) (m map[string]string, errs []error) {
    s := bufio.NewScanner(strings.NewReader(text))

    m = make(map[string]string)
    for lineNum := 1; s.Scan(); lineNum++ {
        key, val, err := ScanKeyValuePair(s.Text())

        if err != nil {
            errs = append(errs, &LinedError{
                Header: "scan config",
                Line: lineNum,
                Err: err,
            })
            continue
        }

        if key == "" && val == "" {
            continue
        }

        m[key] = val
    }
    
    return
}

func ScanInputs(r io.Reader) (inputs Inputs, errs []error) {
    inputs.Config = make(map[string]string)

    s := bufio.NewScanner(r)
    s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
        return splitByString(data, atEOF, TestDelim)
    })

    firstTest := true
    for testId := 1; s.Scan(); testId++ {

        test, testErrs := ScanTest(s.Text())

        if firstTest && testErrs != nil {
            var configErrs []error
            inputs.Config, configErrs = ScanConfig(s.Text())

            if configErrs != nil {
                // Yeah, I know. errs should be empty anyways, so append is
                // redundant. But for the sake of genericity I'll leave it as is.
                errs = append(errs, configErrs...)
                return
            }

            testId--
            firstTest = false
            continue
        }

        firstTest = false

        if testErrs != nil {
            for i, err := range testErrs {
                testErrs[i] = fmt.Errorf("test %d: %w", testId, err)
            }

            errs = append(errs, testErrs...)
            continue
        }

        if test.Input == "" && test.Output == "" {
            testId--
            continue
        }

        inputs.Tests = append(inputs.Tests, test)
    }

    return
}
