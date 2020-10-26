package cptest

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type InputsError string

func (e InputsError) Error() string {
    return string(e)
}

const (
    NoSections = InputsError("IO separator missing")
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
}

func ScanTest(str string) (Test, []error) {
    if strings.TrimSpace(str) == "" {
        return Test{}, nil
    }

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
            return "", "", errors.New("key and value are missing")
        }

        return "", "", errors.New("no equality sign")
    }

    if parts[0] == "" {
        return "", "", errors.New("key cannot be empty")
    }

    if parts[1] == "" {
        return "", "", errors.New("value cannot be empty")
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

func ScanInputs(r io.Reader) (input Inputs, errs []error) {

    s := bufio.NewScanner(r)
    s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
        return splitByString(data, atEOF, TestDelim)
    })

    for testId := 1; s.Scan(); testId++ {
        test, testErrs := ScanTest(s.Text())

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

        input.Tests = append(input.Tests, test)
    }

    return
}
