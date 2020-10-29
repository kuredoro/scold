package cptest

import (
	"bufio"
	"fmt"
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
func ScanInputs(text string) (inputs Inputs, errs []error) {
    // Sentinel test case delimeter
    text += "\n" + TestDelim + "\n"

    inputs.Config = make(map[string]string)

    var str strings.Builder
    testID := 0

    s := bufio.NewScanner(strings.NewReader(text))
    for s.Scan() {

        if strings.HasPrefix(s.Text(), TestDelim) {
            part := str.String()
            str = strings.Builder{}

            test, testErrs := ScanTest(part)

            if test.Input == "" && test.Output == "" && testErrs == nil {
                continue
            }

            // More than 1 configs
            if inputs.Tests == nil && testErrs != nil {
                config, configErrs := ScanConfig(part)

                if configErrs != nil {
                    errs = append(errs, configErrs...)
                }

                for k, v := range config {
                    inputs.Config[k] = v
                }
                continue
            }

            testID++

            if testErrs != nil {
                for i, err := range testErrs {
                    testErrs[i] = fmt.Errorf("test %d: %w", testID, err)
                }
                errs = append(errs, testErrs...)
                continue
            }

            inputs.Tests = append(inputs.Tests, test)
            continue
        }

        line := strings.TrimSpace(s.Text())
        if line != "" {
            str.WriteString(line)
            str.WriteRune('\n')
        }
    }

    return
}
