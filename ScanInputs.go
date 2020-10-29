package cptest

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// InputsError represents an error produced while scanning inputs file.
type InputsError string

func (e InputsError) Error() string {
    return string(e)
}

// A set of errors that may be produced during scanning of the inputs file.
// These replace sentinel errors making the errors be comparable with equals
// operator. 
const (
    IOSeparatorMissing = InputsError("IO separator missing")
    KeyMissing = InputsError("key cannot be empty")
    ValueMissing = InputsError("value cannot be empty")
    KVMissing = InputsError("key and value are missing")
    NotKVPair = InputsError("not a key-value pair")
)

// LinedError appends line information to the error message. It is mainly used
// to test that errors are produced for correct lines.
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


// The set of delimeters used when partitioning inputs file.
const (
    IODelim = "---"
    TestDelim = "==="
)

// Test represents a single test case: an input and the expected output.
type Test struct {
    Input string
    Output string
}

// Inputs contains all information located in the inputs file. It contains
// all of the tests listed there and the set of key-value pairs, if were
// provided.
type Inputs struct {
    Tests []Test
    Config map[string]string
}

// ScanTest parses a single test case: input and output, separated with the
// Input/Output separator. It also trims space around input and output. If
// separator is absent, it returns an error.
func ScanTest(testStr string) (Test, []error) {

    test := Test{}

    var str strings.Builder
    delimFound := false

    s := bufio.NewScanner(strings.NewReader(testStr))
    for s.Scan() {

        if !delimFound && strings.HasPrefix(s.Text(), IODelim) {
            delimFound = true
            test.Input = str.String()
            str = strings.Builder{}
            continue
        }

        line := strings.TrimSpace(s.Text())
        if line != "" {
            str.WriteString(line)
            str.WriteRune('\n')
        }
    }

    if !delimFound && str.String() != "" {
        return Test{}, []error{IOSeparatorMissing}
    }

    test.Output = str.String()

    return test, nil
}

// splitByString is a generic function that splits buffered input by a 
// specified string.
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

// ScanKeyValuePair parses the key-value pair definition of form key=value.
// It returns error if no equality signs are present, or if any side is empty.
// The space around key and value is trimmed.
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

// ScanConfig tries to parse a stream of key-value pairs. It expects each pair
// to be located on a dedicated line. Duplicate keys are allowed, the later
// version is preferred.
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

// ScanInputs is the main routine for parsing inputs file. It splits the input
// by test case separator, and tries to parse each individual test case one by
// one. If the true first test could not be parsed without errors, it is
// interpreted as a configuration and parsed again. The empty tests are
// skipped (those that don't contain input, output and the separator).
// If test case could not be parsed, parsing continues to the next test case,
// but the errors are accumulated and returned together.
func ScanInputs(r io.Reader) (inputs Inputs, errs []error) {
    inputs.Config = make(map[string]string)

    s := bufio.NewScanner(r)
    s.Split(func(data []byte, atEOF bool) (int, []byte, error) {
        return splitByString(data, atEOF, TestDelim)
    })

    firstTest := true
    for testID := 1; s.Scan(); testID++ {

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

            testID--
            firstTest = false
            continue
        }

        firstTest = false

        if testErrs != nil {
            for i, err := range testErrs {
                testErrs[i] = fmt.Errorf("test %d: %w", testID, err)
            }

            errs = append(errs, testErrs...)
            continue
        }

        if test.Input == "" && test.Output == "" {
            testID--
            continue
        }

        inputs.Tests = append(inputs.Tests, test)
    }

    return
}
